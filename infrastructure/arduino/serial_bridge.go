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
	mu              sync.Mutex
}

func NewSerialBridge(hub *websocket.Hub) *SerialBridge {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SerialBridge{
		config: SerialConfig{
			BaudRate:       115200,
			DataBits:       8,
			StopBits:       1,
			Parity:         "N",
			ReadTimeout:    time.Second * 30,
			ReconnectDelay: time.Second * 5,
			MaxReconnects:  10,
			AutoRestore:    true,
		},
		devices:         make(map[string]*DeviceInfo),
		restoredDevices: make(map[string]bool),
		hub:             hub,
		ctx:             ctx,
		cancel:          cancel,
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
	
	sb.port = port
	sb.lastPort = portPath
	sb.reconnectCount = 0
	sb.reconnecting = false
	
	log.Printf("✅ Conectado a Arduino en %s", portPath)
	
	go sb.readLoop()
	
	time.Sleep(3 * time.Second)
	sb.restoreAllDevices()
	
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
				log.Printf("❌ Error leyendo puerto: %v", err)
				continue
			}
			
			sb.processLine(line)
		}
	}
}

func (sb *SerialBridge) processLine(line string) {
	line = strings.TrimSpace(line)
	
	if line == "" || !strings.HasPrefix(line, "{") {
		if line != "" {
			log.Printf("Arduino: %s", line)
		}
		return
	}
	
	var data ArduinoData
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		log.Printf("⚠️ JSON inválido: %s", line[:min(100, len(line))])
		return
	}
	
	if data.Type != "data" {
		return
	}
	
	log.Printf("📊 Datos de %s: Power=%.2fW Energy=%.3fkWh Servicio=%v", 
		data.DeviceID, data.Power, data.Energy, data.ServicioActivo)
	
	sb.devicesMu.Lock()
	device, exists := sb.devices[data.DeviceID]
	if !exists {
		device = &DeviceInfo{
			ID: data.DeviceID,
		}
		sb.devices[data.DeviceID] = device
		sb.devicesMu.Unlock()
		
		go sb.registerDevice(&data)
		
		if !sb.restoredDevices[data.DeviceID] {
			sb.restoredDevices[data.DeviceID] = true
			go sb.restoreDeviceState(data.DeviceID)
		}
	} else {
		sb.devicesMu.Unlock()
	}
	
	device.LastReading = &data
	
	go sb.saveReading(&data)
	
	sb.sendToWebSocket(&data)
}

func (sb *SerialBridge) saveReading(data *ArduinoData) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	update := bson.M{
		"$set": bson.M{
			"ultimaLectura": bson.M{
				"voltage":     data.Voltage,
				"current":     data.Current,
				"activePower": data.Power,
				"energy":      data.Energy,
				"cost":        data.Cost,
				"timestamp":   time.Now(),
			},
			"estado": "activo",
		},
	}
	
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"numeroDispositivo": data.DeviceID},
		update,
	)
	
	if err != nil {
		log.Printf("❌ Error guardando lectura de %s: %v", data.DeviceID, err)
	}
}

func (sb *SerialBridge) registerDevice(data *ArduinoData) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var dispositivo bson.M
	err := collection.FindOne(ctx, bson.M{"numeroDispositivo": data.DeviceID}).Decode(&dispositivo)
	
	if err == nil {
		log.Printf("✅ Dispositivo %s ya registrado", data.DeviceID)
		
		sb.devicesMu.Lock()
		if device, ok := sb.devices[data.DeviceID]; ok {
			if clienteID, ok := dispositivo["clienteAsignado"].(string); ok {
				device.ClienteID = clienteID
			}
			if empresaID, ok := dispositivo["empresaAsignada"].(string); ok {
				device.EmpresaID = empresaID
			}
		}
		sb.devicesMu.Unlock()
		return
	}
	
	nuevoDispositivo := bson.M{
		"numeroDispositivo": data.DeviceID,
		"nombre":           fmt.Sprintf("Arduino %s", data.DeviceID),
		"tipo":             "arduino_uno",
		"estado":           "activo",
		"direccion":        "Sin asignar",
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
		"idDispositivo":   data.DeviceID,
		"potenciaActiva":  data.Power,
		"energia":         data.Energy,
		"voltaje":         data.Voltage,
		"corriente":       data.Current,
		"costo":           data.Cost,
		"marcaTiempo":     time.Now().Format(time.RFC3339),
		"servicioActivo":  data.ServicioActivo,
		"uptime":          data.Uptime,
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
	
	if ultimaLectura, ok := dispositivo["ultimaLectura"].(bson.M); ok {
		if energia, ok := ultimaLectura["energia"].(float64); ok {
			costo := 0.0
			if c, ok := ultimaLectura["costo"].(float64); ok {
				costo = c
			}
			
			comando := fmt.Sprintf("RESTORE:%.3f:%.2f", energia, costo)
			sb.SendCommand(comando)
			
			log.Printf("✅ Estado restaurado para %s: Energía=%.3fkWh Costo=%.2f", 
				deviceID, energia, costo)
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
