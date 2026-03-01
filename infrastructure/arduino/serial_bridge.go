package arduino

import (
	"bufio"
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/websocket"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SerialBridge struct {
	port            serial.Port
	config          SerialConfig
	devices         map[string]*DeviceInfo
	devicesMu       sync.RWMutex
	hub             *websocket.Hub
	ctx             context.Context
	cancel          context.CancelFunc
	reconnecting    bool
	reconnectCount  int
	lastPort        string
	restoredDevices map[string]bool
	aggregator      *ReadingAggregator
	mu              sync.Mutex
}

func NewSerialBridge(hub *websocket.Hub) *SerialBridge {
	ctx, cancel := context.WithCancel(context.Background())

	aggregator := GetAggregator()
	if err := aggregator.Initialize(); err != nil {
		log.Printf("⚠️ Error inicializando agregador: %v", err)
	}

	return &SerialBridge{
		config: SerialConfig{
			BaudRate:       115200,
			DataBits:       8,
			StopBits:       1,
			Parity:         "N",
			ReadTimeout:    time.Millisecond * 500,
			ReconnectDelay: time.Second * 5,
			MaxReconnects:  10,
			AutoRestore:    true,
		},
		devices:         make(map[string]*DeviceInfo),
		restoredDevices: make(map[string]bool),
		hub:             hub,
		ctx:             ctx,
		cancel:          cancel,
		aggregator:      aggregator,
	}
}

func (sb *SerialBridge) ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	log.Printf("📋 Puertos seriales disponibles:")
	for i, port := range ports {
		log.Printf("  %d. %s", i+1, port)
	}

	return ports, nil
}

func (sb *SerialBridge) findArduinoPort() (string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return "", err
	}

	for _, port := range ports {
		if strings.Contains(port, "ttyUSB") ||
			strings.Contains(port, "ttyACM") ||
			strings.Contains(port, "COM") {
			if !strings.Contains(port, "ttyS") {
				log.Printf("✅ Arduino encontrado en %s", port)
				return port, nil
			}
		}
	}

	return "", fmt.Errorf("no se encontró Arduino conectado")
}

func (sb *SerialBridge) Connect(portPath string) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if portPath == "" {
		foundPort, err := sb.findArduinoPort()
		if err != nil {
			return err
		}
		portPath = foundPort
	}

	log.Printf("🔌 Conectando a Arduino en %s...", portPath)

	mode := &serial.Mode{
		BaudRate: sb.config.BaudRate,
		DataBits: sb.config.DataBits,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}

	port, err := serial.Open(portPath, mode)
	if err != nil {
		return fmt.Errorf("error abriendo puerto: %w", err)
	}

	// Timeout corto para que el readLoop no se bloquee en Windows
	port.SetReadTimeout(sb.config.ReadTimeout)

	sb.port = port
	sb.lastPort = portPath
	sb.reconnectCount = 0
	sb.reconnecting = false

	log.Printf("✅ Conectado a Arduino en %s", portPath)

	go sb.readLoop()

	return nil
}

func (sb *SerialBridge) readLoop() {
	reader := bufio.NewReader(sb.port)

	for {
		select {
		case <-sb.ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "file already closed") {
					log.Printf("⚠️ Puerto serial cerrado")
					sb.handleDisconnection()
					return
				}
				// En Windows, timeout de lectura es normal cuando no hay datos — ignorar
				if strings.Contains(err.Error(), "multiple Read calls") ||
					strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "i/o timeout") {
					// Si hay datos parciales, procesarlos igual
					if len(line) > 0 {
						sb.processLine(line)
					}
					continue
				}
				log.Printf("❌ Error leyendo puerto: %v", err)
				continue
			}

			sb.processLine(line)
		}
	}
}

func (sb *SerialBridge) processLine(line string) {
	line = strings.TrimSpace(line)

	// Limpiar caracteres no imprimibles / BOM al inicio
	for len(line) > 0 && (line[0] < 0x20 || line[0] > 0x7E) {
		line = line[1:]
	}

	if line == "" {
		return
	}

	// Buscar inicio del JSON si hay basura antes
	if idx := strings.Index(line, "{"); idx > 0 {
		line = line[idx:]
	}

	if !strings.HasPrefix(line, "{") {
		return
	}

	var data ArduinoData
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return
	}

	if data.Type != "data" {
		return
	}

	sb.devicesMu.Lock()
	device, exists := sb.devices[data.DeviceID]
	if !exists {
		device = &DeviceInfo{
			ID: data.DeviceID,
		}
		sb.devices[data.DeviceID] = device
		sb.devicesMu.Unlock()

		// Registrar/cargar dispositivo de forma SÍNCRONA para tener ClienteID antes de enviar WS
		sb.registerDevice(&data)

		if !sb.restoredDevices[data.DeviceID] {
			sb.restoredDevices[data.DeviceID] = true
			go sb.restoreDeviceState(data.DeviceID)
		}
	} else {
		// Si ya existe pero no tiene ClienteID cargado, restaurar desde DB
		needsRestore := device.ClienteID == ""
		sb.devicesMu.Unlock()
		if needsRestore && !sb.restoredDevices[data.DeviceID] {
			sb.restoredDevices[data.DeviceID] = true
			sb.restoreDeviceState(data.DeviceID) // síncrono también
		}
	}

	device.LastReading = &data

	sb.devicesMu.RLock()
	clienteID := ""
	empresaID := ""
	if d, ok := sb.devices[data.DeviceID]; ok {
		clienteID = d.ClienteID
		empresaID = d.EmpresaID
	}
	sb.devicesMu.RUnlock()

	sb.aggregator.AddReading(&data, clienteID, empresaID)

	go sb.saveReading(&data)

	sb.sendToWebSocket(&data)
}

func (sb *SerialBridge) saveReading(data *ArduinoData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispositivosCol := config.MongoDB.Collection("dispositivos")
	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"ultimaLectura": bson.M{
				"voltage":     data.Voltage,
				"current":     data.Current,
				"activePower": data.Power,
				"energy":      data.Energy,
				"cost":        data.Cost,
				"timestamp":   now,
			},
			"estado": "activo",
		},
	}

	_, err := dispositivosCol.UpdateOne(ctx, bson.M{"numeroDispositivo": data.DeviceID}, update)
	if err != nil {
		log.Printf("❌ Error guardando lectura de %s: %v", data.DeviceID, err)
	}

	// Obtener clienteId del dispositivo para guardar historial
	var dispositivo bson.M
	err = dispositivosCol.FindOne(ctx, bson.M{"numeroDispositivo": data.DeviceID}).Decode(&dispositivo)
	if err != nil {
		return
	}

	clienteId, _ := dispositivo["clienteId"]
	dispositivoId := dispositivo["_id"]

	lecturasCol := config.MongoDB.Collection("lecturas")
	lectura := bson.M{
		"timestamp":     now,
		"dispositivoId": dispositivoId,
		"clienteId":     clienteId,
		"voltaje":       data.Voltage,
		"corriente":     data.Current,
		"potencia":      data.Power,
		"energia":       data.Energy,
		"costo":         data.Cost,
	}
	if _, err := lecturasCol.InsertOne(ctx, lectura); err != nil {
		log.Printf("⚠️ Error guardando historial lectura: %v", err)
	}
}

func (sb *SerialBridge) registerDevice(data *ArduinoData) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dispositivo bson.M
	err := collection.FindOne(ctx, bson.M{"numeroDispositivo": data.DeviceID}).Decode(&dispositivo)

	if err == nil {
		sb.devicesMu.Lock()
		if device, ok := sb.devices[data.DeviceID]; ok {
			if clienteOID, ok := dispositivo["clienteId"].(primitive.ObjectID); ok {
				device.ClienteID = clienteOID.Hex()
			} else if clienteStr, ok := dispositivo["clienteId"].(string); ok {
				device.ClienteID = clienteStr
			}
			if empresaOID, ok := dispositivo["empresaId"].(primitive.ObjectID); ok {
				device.EmpresaID = empresaOID.Hex()
			}
		}
		sb.devicesMu.Unlock()
		return
	}

	nuevoDispositivo := bson.M{
		"numeroDispositivo": data.DeviceID,
		"nombre":            fmt.Sprintf("Arduino %s", data.DeviceID),
		"tipo":              "arduino_uno",
		"estado":            "activo",
		"direccion":         "Sin asignar",
		"configuracion": bson.M{
			"voltajeNominal":  220,
			"corrienteMaxima": 0.5,
			"potenciaMaxima":  110,
			"tarifaKwh":       150,
		},
		"activo":        true,
		"fechaCreacion": time.Now().UnixMilli(),
	}

	_, err = collection.InsertOne(ctx, nuevoDispositivo)
	if err != nil {
		log.Printf("❌ Error registrando dispositivo: %v", err)
		return
	}

	log.Printf("✅ Dispositivo %s registrado exitosamente (sin asignar)", data.DeviceID)
}

func (sb *SerialBridge) sendToWebSocket(data *ArduinoData) {
	sb.devicesMu.RLock()
	device := sb.devices[data.DeviceID]
	sb.devicesMu.RUnlock()

	if device == nil {
		return
	}

	wsData := map[string]interface{}{
		"idDispositivo":  data.DeviceID,
		"potenciaActiva": data.Power,
		"energia":        data.Energy,
		"voltaje":        data.Voltage,
		"corriente":      data.Current,
		"costo":          data.Cost,
		"marcaTiempo":    time.Now().Format(time.RFC3339),
		"servicioActivo": data.ServicioActivo,
		"uptime":         data.Uptime,
	}

	msg := websocket.Message{
		Type:      websocket.MessageTypeDeviceUpdate,
		Data:      wsData,
		Timestamp: time.Now(),
		ClienteID: device.ClienteID,
	}

	if device.ClienteID != "" {
		sb.hub.BroadcastToCliente(device.ClienteID, msg)
	}

	if device.EmpresaID != "" {
		sb.hub.BroadcastToEmpresa(device.EmpresaID, msg)
	}
}

func (sb *SerialBridge) restoreAllDevices() {
	sb.devicesMu.RLock()
	devices := make([]string, 0, len(sb.devices))
	for deviceID := range sb.devices {
		devices = append(devices, deviceID)
	}
	sb.devicesMu.RUnlock()

	log.Printf("🔄 Restaurando %d dispositivo(s)...", len(devices))

	for _, deviceID := range devices {
		if !sb.restoredDevices[deviceID] {
			sb.restoredDevices[deviceID] = true
			sb.restoreDeviceState(deviceID)
		}
	}
}

func (sb *SerialBridge) restoreDeviceState(deviceID string) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dispositivo bson.M
	err := collection.FindOne(ctx, bson.M{"numeroDispositivo": deviceID}).Decode(&dispositivo)

	if err != nil {
		log.Printf("⚠️ No se pudo restaurar estado de %s", deviceID)
		return
	}

	sb.devicesMu.Lock()
	if device, ok := sb.devices[deviceID]; ok {
		if clienteOID, ok := dispositivo["clienteId"].(primitive.ObjectID); ok {
			device.ClienteID = clienteOID.Hex()
		} else if clienteStr, ok := dispositivo["clienteId"].(string); ok {
			device.ClienteID = clienteStr
		}
		if empresaOID, ok := dispositivo["empresaId"].(primitive.ObjectID); ok {
			device.EmpresaID = empresaOID.Hex()
		}
	}
	sb.devicesMu.Unlock()

	if ultimaLectura, ok := dispositivo["ultimaLectura"].(bson.M); ok {
		if energia, ok := ultimaLectura["energia"].(float64); ok {
			costo := 0.0
			if c, ok := ultimaLectura["costo"].(float64); ok {
				costo = c
			}
			comando := fmt.Sprintf("RESTORE:%.3f:%.2f", energia, costo)
			sb.SendCommand(comando)
		}
	}
}

func (sb *SerialBridge) SendCommand(command string) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.port == nil {
		return fmt.Errorf("puerto no conectado")
	}

	log.Printf("📤 Enviando comando: %s", command)

	_, err := sb.port.Write([]byte(command + "\n"))
	if err != nil {
		log.Printf("❌ Error enviando comando: %v", err)
		return err
	}

	log.Printf("✅ Comando enviado: %s", command)
	return nil
}

func (sb *SerialBridge) handleDisconnection() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	log.Printf("⚠️ Arduino desconectado")

	sb.saveAllDevicesState()

	for deviceID := range sb.restoredDevices {
		delete(sb.restoredDevices, deviceID)
	}

	if sb.port != nil {
		sb.port.Close()
		sb.port = nil
	}

	sb.scheduleReconnect()
}

func (sb *SerialBridge) saveAllDevicesState() {
	sb.devicesMu.RLock()
	defer sb.devicesMu.RUnlock()

	log.Printf("💾 Guardando estado de %d dispositivo(s)...", len(sb.devices))

	collection := config.MongoDB.Collection("dispositivos")

	for deviceID, device := range sb.devices {
		if device.LastReading == nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		update := bson.M{
			"$set": bson.M{
				"ultimaLectura": bson.M{
					"voltaje":   device.LastReading.Voltage,
					"corriente": device.LastReading.Current,
					"potencia":  device.LastReading.Power,
					"energia":   device.LastReading.Energy,
					"costo":     device.LastReading.Cost,
					"timestamp": time.Now(),
				},
			},
		}

		_, err := collection.UpdateOne(ctx, bson.M{"numeroDispositivo": deviceID}, update)
		cancel()

		if err != nil {
			log.Printf("❌ Error guardando estado de %s: %v", deviceID, err)
		} else {
			log.Printf("✅ Estado guardado para %s", deviceID)
		}
	}
}

func (sb *SerialBridge) scheduleReconnect() {
	if sb.reconnecting || sb.reconnectCount >= sb.config.MaxReconnects {
		if sb.reconnectCount >= sb.config.MaxReconnects {
			log.Printf("❌ Máximo de reconexiones alcanzado (%d)", sb.config.MaxReconnects)
		}
		return
	}

	sb.reconnecting = true
	sb.reconnectCount++

	log.Printf("🔄 Programando reconexión (intento %d/%d) en %v...",
		sb.reconnectCount, sb.config.MaxReconnects, sb.config.ReconnectDelay)

	time.AfterFunc(sb.config.ReconnectDelay, func() {
		sb.attemptReconnect()
	})
}

func (sb *SerialBridge) attemptReconnect() {
	log.Printf("🔌 Intentando reconectar (intento %d/%d)...",
		sb.reconnectCount, sb.config.MaxReconnects)

	portToUse, err := sb.findArduinoPort()
	if err != nil {
		log.Printf("⚠️ No se encontró Arduino: %v", err)
		sb.reconnecting = false
		sb.scheduleReconnect()
		return
	}

	sb.mu.Lock()
	sb.reconnecting = false
	sb.mu.Unlock()

	if err := sb.Connect(portToUse); err != nil {
		log.Printf("❌ Error en reconexión: %v", err)
		sb.scheduleReconnect()
	} else {
		log.Printf("✅ Reconexión exitosa")
	}
}

func (sb *SerialBridge) IsConnected() bool {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.port != nil
}

func (sb *SerialBridge) GetDevices() []*DeviceInfo {
	sb.devicesMu.RLock()
	defer sb.devicesMu.RUnlock()

	devices := make([]*DeviceInfo, 0, len(sb.devices))
	for _, device := range sb.devices {
		devices = append(devices, device)
	}
	return devices
}

func (sb *SerialBridge) Disconnect() {
	sb.cancel()

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.aggregator != nil {
		sb.aggregator.FlushAll()
	}

	if sb.port != nil {
		sb.saveAllDevicesState()
		sb.port.Close()
		sb.port = nil
		log.Printf("✅ Puerto serial cerrado")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
