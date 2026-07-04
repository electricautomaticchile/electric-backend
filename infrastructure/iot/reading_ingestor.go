package iot

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Reading struct {
	DeviceID   string
	Lectura    *entities.LecturaDispositivo
	ReceivedAt time.Time
	// Latitud/Longitud son opcionales: solo se persisten como última ubicación
	// conocida del dispositivo cuando el ESP32 reporta GPS en la lectura.
	Latitud  *float64
	Longitud *float64
}

type ReadingIngestor struct {
	collection    *mongo.Collection
	queue         chan Reading
	batchSize     int
	workers       int
	flushInterval time.Duration
	writeTimeout  time.Duration
	maxRetries    int
	stop          chan struct{}
	stopOnce      sync.Once
	wg            sync.WaitGroup
	enqueued      atomic.Uint64
	flushed       atomic.Uint64
	dropped       atomic.Uint64
	failed        atomic.Uint64
}

var (
	defaultMu       sync.RWMutex
	defaultIngestor *ReadingIngestor
)

func StartDefaultReadingIngestor(db *mongo.Database, cfg *config.Config) {
	if db == nil || cfg == nil || !cfg.IoTIngestAsync {
		return
	}

	ingestor := NewReadingIngestor(
		db.Collection(entities.DispositivoEntity{}.CollectionName()),
		cfg.IoTIngestQueueSize,
		cfg.IoTIngestBatchSize,
		cfg.IoTIngestWorkers,
		cfg.IoTIngestFlushInterval,
		cfg.IoTIngestWriteTimeout,
		cfg.IoTIngestMaxRetries,
	)
	ingestor.Start()

	defaultMu.Lock()
	defaultIngestor = ingestor
	defaultMu.Unlock()

	log.Printf(
		"✅ IoT async ingestor activo (queue=%d batch=%d workers=%d flush=%s write_timeout=%s retries=%d)",
		cfg.IoTIngestQueueSize,
		cfg.IoTIngestBatchSize,
		cfg.IoTIngestWorkers,
		cfg.IoTIngestFlushInterval,
		cfg.IoTIngestWriteTimeout,
		cfg.IoTIngestMaxRetries,
	)
}

func StopDefaultReadingIngestor(ctx context.Context) {
	defaultMu.RLock()
	ingestor := defaultIngestor
	defaultMu.RUnlock()
	if ingestor == nil {
		return
	}
	ingestor.Stop(ctx)
}

func DefaultReadingIngestor() *ReadingIngestor {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultIngestor
}

func NewReadingIngestor(collection *mongo.Collection, queueSize int, batchSize int, workers int, flushInterval time.Duration, writeTimeout time.Duration, maxRetries int) *ReadingIngestor {
	if queueSize < 1 {
		queueSize = 100000
	}
	if batchSize < 1 {
		batchSize = 1000
	}
	if workers < 1 {
		workers = 1
	}
	if flushInterval <= 0 {
		flushInterval = 100 * time.Millisecond
	}
	if writeTimeout <= 0 {
		writeTimeout = 15 * time.Second
	}
	if maxRetries < 1 {
		maxRetries = 1
	}

	return &ReadingIngestor{
		collection:    collection,
		queue:         make(chan Reading, queueSize),
		batchSize:     batchSize,
		workers:       workers,
		flushInterval: flushInterval,
		writeTimeout:  writeTimeout,
		maxRetries:    maxRetries,
		stop:          make(chan struct{}),
	}
}

func (i *ReadingIngestor) Start() {
	for worker := 0; worker < i.workers; worker++ {
		i.wg.Add(1)
		go i.runWorker()
	}
}

func (i *ReadingIngestor) Stop(ctx context.Context) {
	i.stopOnce.Do(func() {
		close(i.stop)
	})

	done := make(chan struct{})
	go func() {
		i.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Printf("⚠️ IoT ingestor no alcanzo a vaciar cola: %v", ctx.Err())
	}
}

func (i *ReadingIngestor) Enqueue(reading Reading) bool {
	if reading.DeviceID == "" || reading.Lectura == nil {
		i.dropped.Add(1)
		return false
	}
	if reading.ReceivedAt.IsZero() {
		reading.ReceivedAt = time.Now().UTC()
	}

	select {
	case i.queue <- reading:
		i.enqueued.Add(1)
		return true
	default:
		i.dropped.Add(1)
		return false
	}
}

func (i *ReadingIngestor) Stats() map[string]uint64 {
	return map[string]uint64{
		"enqueued": i.enqueued.Load(),
		"flushed":  i.flushed.Load(),
		"dropped":  i.dropped.Load(),
		"failed":   i.failed.Load(),
		"queued":   uint64(len(i.queue)),
	}
}

func (i *ReadingIngestor) runWorker() {
	defer i.wg.Done()

	ticker := time.NewTicker(i.flushInterval)
	defer ticker.Stop()

	pending := make(map[string]Reading, i.batchSize)
	for {
		select {
		case reading := <-i.queue:
			pending[reading.DeviceID] = reading
			if len(pending) >= i.batchSize {
				i.flush(pending)
				pending = make(map[string]Reading, i.batchSize)
			}
		case <-ticker.C:
			if len(pending) > 0 {
				i.flush(pending)
				pending = make(map[string]Reading, i.batchSize)
			}
		case <-i.stop:
			i.drainInto(pending)
			if len(pending) > 0 {
				i.flush(pending)
			}
			return
		}
	}
}

func (i *ReadingIngestor) drainInto(pending map[string]Reading) {
	for {
		select {
		case reading := <-i.queue:
			pending[reading.DeviceID] = reading
		default:
			return
		}
	}
}

func (i *ReadingIngestor) flush(pending map[string]Reading) {
	if len(pending) == 0 {
		return
	}

	models := make([]mongo.WriteModel, 0, len(pending))
	for _, reading := range pending {
		set := bson.M{
			"ultimaLectura":      reading.Lectura,
			"fechaActualizacion": reading.ReceivedAt,
		}
		// Persistir última ubicación conocida solo si la lectura trae GPS válido.
		if reading.Latitud != nil && reading.Longitud != nil {
			set["latitud"] = *reading.Latitud
			set["longitud"] = *reading.Longitud
		}

		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"numeroDispositivo": reading.DeviceID,
				"activo":            true,
			}).
			SetUpdate(bson.M{
				"$set": set,
			}),
		)
	}

	var lastErr error
	for attempt := 1; attempt <= i.maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), i.writeTimeout)
		_, err := i.collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
		cancel()
		if err == nil {
			i.flushed.Add(uint64(len(models)))
			return
		}
		lastErr = err
		if attempt < i.maxRetries {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}

	i.failed.Add(uint64(len(models)))
	log.Printf("⚠️ error escribiendo batch IoT tras %d intento(s) (%d lecturas): %v", i.maxRetries, len(models), lastErr)
}
