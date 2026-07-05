package arduino

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeviceBuffer struct {
	DispositivoID string
	ClienteID     string
	EmpresaID     string
	Lecturas      []ArduinoData
	mu            sync.Mutex
}

type ReadingAggregator struct {
	buffers        map[string]*DeviceBuffer
	mu             sync.RWMutex
	flushInterval  time.Duration
	collection     *mongo.Collection
	collectionAgg  *mongo.Collection
	ctx            context.Context
	cancel         context.CancelFunc
	initialized    bool
}

var (
	aggregatorInstance *ReadingAggregator
	aggregatorOnce     sync.Once
)

func GetAggregator() *ReadingAggregator {
	aggregatorOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		aggregatorInstance = &ReadingAggregator{
			buffers:       make(map[string]*DeviceBuffer),
			flushInterval: 5 * time.Minute,
			ctx:           ctx,
			cancel:        cancel,
		}
	})
	return aggregatorInstance
}

func (ra *ReadingAggregator) Initialize() error {
	if ra.initialized {
		return nil
	}

	ra.collection = config.MongoDB.Collection(entities.LecturaEntity{}.CollectionName())
	ra.collectionAgg = config.MongoDB.Collection(entities.LecturaAgregadaEntity{}.CollectionName())

	// La colección cruda 'lecturas' quedó en desuso: la ingesta 4G solo actualiza
	// dispositivos.ultimaLectura y el histórico se guarda agregado en
	// 'lecturas_agregadas'. Por eso ya no creamos la time-series 'lecturas'.

	if err := ra.ensureIndexes(); err != nil {
		log.Printf("⚠️ Error creando índices: %v", err)
	}

	go ra.flushLoop()

	ra.initialized = true
	log.Printf("✅ Agregador de lecturas inicializado (flush cada %v)", ra.flushInterval)
	return nil
}

func (ra *ReadingAggregator) ensureTimeSeriesCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections, err := config.MongoDB.ListCollectionNames(ctx, bson.M{"name": "lecturas"})
	if err != nil {
		return err
	}

	if len(collections) > 0 {
		log.Printf("✅ Colección 'lecturas' ya existe")
		return nil
	}

	opts := options.CreateCollection().
		SetTimeSeriesOptions(&options.TimeSeriesOptions{
			TimeField:   "timestamp",
			MetaField:   stringPtr("dispositivoId"),
			Granularity: stringPtr("minutes"),
		}).
		SetExpireAfterSeconds(7776000)

	err = config.MongoDB.CreateCollection(ctx, "lecturas", opts)
	if err != nil {
		return err
	}

	log.Printf("✅ Time Series Collection 'lecturas' creada (TTL: 90 días)")
	return nil
}

func (ra *ReadingAggregator) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "dispositivoId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "clienteId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "empresaId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
	}

	_, err := ra.collectionAgg.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(31536000),
	}
	_, err = ra.collectionAgg.Indexes().CreateOne(ctx, ttlIndex)

	log.Printf("✅ Índices creados para lecturas_agregadas (TTL: 365 días)")
	return err
}

func (ra *ReadingAggregator) AddReading(data *ArduinoData, clienteID, empresaID string) {
	ra.mu.Lock()
	buffer, exists := ra.buffers[data.DeviceID]
	if !exists {
		buffer = &DeviceBuffer{
			DispositivoID: data.DeviceID,
			ClienteID:     clienteID,
			EmpresaID:     empresaID,
			Lecturas:      make([]ArduinoData, 0, 100),
		}
		ra.buffers[data.DeviceID] = buffer
	}
	ra.mu.Unlock()

	buffer.mu.Lock()
	buffer.Lecturas = append(buffer.Lecturas, *data)
	if clienteID != "" {
		buffer.ClienteID = clienteID
	}
	if empresaID != "" {
		buffer.EmpresaID = empresaID
	}
	buffer.mu.Unlock()
}

func (ra *ReadingAggregator) flushLoop() {
	ticker := time.NewTicker(ra.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ra.ctx.Done():
			ra.FlushAll()
			return
		case <-ticker.C:
			ra.FlushAll()
		}
	}
}

func (ra *ReadingAggregator) FlushAll() {
	ra.mu.RLock()
	deviceIDs := make([]string, 0, len(ra.buffers))
	for id := range ra.buffers {
		deviceIDs = append(deviceIDs, id)
	}
	ra.mu.RUnlock()

	for _, deviceID := range deviceIDs {
		ra.flushDevice(deviceID)
	}
}

func (ra *ReadingAggregator) flushDevice(deviceID string) {
	ra.mu.RLock()
	buffer, exists := ra.buffers[deviceID]
	ra.mu.RUnlock()

	if !exists {
		return
	}

	buffer.mu.Lock()
	if len(buffer.Lecturas) == 0 {
		buffer.mu.Unlock()
		return
	}

	lecturas := buffer.Lecturas
	buffer.Lecturas = make([]ArduinoData, 0, 100)
	clienteID := buffer.ClienteID
	empresaID := buffer.EmpresaID
	buffer.mu.Unlock()

	aggregated := ra.aggregate(deviceID, clienteID, empresaID, lecturas)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ra.collectionAgg.InsertOne(ctx, aggregated)
	if err != nil {
		log.Printf("❌ Error guardando agregación de %s: %v", deviceID, err)
		return
	}

	log.Printf("✅ Agregación guardada: %s (%d lecturas, Potencia avg: %.2fW)", 
		deviceID, aggregated.NumLecturas, aggregated.PotenciaAvg)
}

func (ra *ReadingAggregator) aggregate(deviceID, clienteID, empresaID string, lecturas []ArduinoData) *entities.LecturaAgregadaEntity {
	if len(lecturas) == 0 {
		return nil
	}

	var voltajeSum, corrienteSum, potenciaSum float64
	voltajeMin, voltajeMax := lecturas[0].Voltage, lecturas[0].Voltage
	corrienteMin, corrienteMax := lecturas[0].Current, lecturas[0].Current
	potenciaMin, potenciaMax := lecturas[0].Power, lecturas[0].Power

	for _, l := range lecturas {
		voltajeSum += l.Voltage
		corrienteSum += l.Current
		potenciaSum += l.Power

		if l.Voltage < voltajeMin {
			voltajeMin = l.Voltage
		}
		if l.Voltage > voltajeMax {
			voltajeMax = l.Voltage
		}
		if l.Current < corrienteMin {
			corrienteMin = l.Current
		}
		if l.Current > corrienteMax {
			corrienteMax = l.Current
		}
		if l.Power < potenciaMin {
			potenciaMin = l.Power
		}
		if l.Power > potenciaMax {
			potenciaMax = l.Power
		}
	}

	n := float64(len(lecturas))

	return &entities.LecturaAgregadaEntity{
		Timestamp:     time.Now(),
		DispositivoID: deviceID,
		ClienteID:     clienteID,
		EmpresaID:     empresaID,
		Periodo:       "5min",
		VoltajeMin:    voltajeMin,
		VoltajeMax:    voltajeMax,
		VoltajeAvg:    voltajeSum / n,
		CorrienteMin:  corrienteMin,
		CorrienteMax:  corrienteMax,
		CorrienteAvg:  corrienteSum / n,
		PotenciaMin:   potenciaMin,
		PotenciaMax:   potenciaMax,
		PotenciaAvg:   potenciaSum / n,
		EnergiaInicio: lecturas[0].Energy,
		EnergiaFin:    lecturas[len(lecturas)-1].Energy,
		CostoInicio:   lecturas[0].Cost,
		CostoFin:      lecturas[len(lecturas)-1].Cost,
		NumLecturas:   len(lecturas),
	}
}

func (ra *ReadingAggregator) Stop() {
	ra.cancel()
}

func stringPtr(s string) *string {
	return &s
}
