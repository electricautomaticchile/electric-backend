package arduino

import (
	"bufio"
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/eventbus"
	"electric-backend/infrastructure/sms"
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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SerialBridge struct {
	ports           map[string]serial.Port // portPath → puerto abierto
	config          SerialConfig
	devices         map[string]*DeviceInfo
	devicesMu       sync.RWMutex
	publisher       *eventbus.Publisher
	ctx             context.Context
	cancel          context.CancelFunc
	restoredDevices map[string]bool
	aggregator      *ReadingAggregator
	mu              sync.Mutex
	// Umbrales de notificación por costo
	costThresholds   map[string]float64 // deviceID → último umbral notificado
	costThresholdsMu sync.Mutex
}

func NewSerialBridge(publisher *eventbus.Publisher) *SerialBridge {
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
		ports:           make(map[string]serial.Port),
		devices:         make(map[string]*DeviceInfo),
		restoredDevices: make(map[string]bool),
		costThresholds:  make(map[string]float64),
		publisher:       publisher,
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

func (sb *SerialBridge) findAllArduinoPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	var found []string
	for _, port := range ports {
		if (strings.Contains(port, "ttyUSB") ||
			strings.Contains(port, "ttyACM") ||
			strings.Contains(port, "COM")) &&
			!strings.Contains(port, "ttyS") {
			found = append(found, port)
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no se encontraron Arduinos conectados")
	}
	return found, nil
}

// Connect conecta un puerto específico (o auto-detecta si portPath=="")
func (sb *SerialBridge) Connect(portPath string) error {
	if portPath == "" {
		ports, err := sb.findAllArduinoPorts()
		if err != nil {
			return err
		}
		for _, p := range ports {
			go sb.connectPort(p)
		}
		return nil
	}
	return sb.connectPort(portPath)
}

func (sb *SerialBridge) connectPort(portPath string) error {
	sb.mu.Lock()
	if _, already := sb.ports[portPath]; already {
		sb.mu.Unlock()
		return nil // ya conectado
	}
	sb.mu.Unlock()

	log.Printf("🔌 Conectando a Arduino en %s...", portPath)
	mode := &serial.Mode{
		BaudRate: sb.config.BaudRate,
		DataBits: sb.config.DataBits,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}
	port, err := serial.Open(portPath, mode)
	if err != nil {
		return fmt.Errorf("error abriendo %s: %w", portPath, err)
	}
	port.SetReadTimeout(sb.config.ReadTimeout)

	sb.mu.Lock()
	sb.ports[portPath] = port
	sb.mu.Unlock()

	log.Printf("✅ Conectado a Arduino en %s", portPath)
	go sb.readLoop(portPath, port)
	return nil
}

func (sb *SerialBridge) readLoop(portPath string, port serial.Port) {
	reader := bufio.NewReader(port)
	for {
		select {
		case <-sb.ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "file already closed") {
					log.Printf("⚠️ Puerto %s cerrado — reconectando...", portPath)
					sb.handlePortDisconnection(portPath)
					return
				}
				if strings.Contains(err.Error(), "multiple Read calls") ||
					strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "i/o timeout") {
					if len(line) > 0 {
						sb.processLine(line, portPath)
					}
					continue
				}
				log.Printf("❌ Error leyendo %s: %v", portPath, err)
				continue
			}
			sb.processLine(line, portPath)
		}
	}
}

func (sb *SerialBridge) handlePortDisconnection(portPath string) {
	sb.mu.Lock()
	if port, ok := sb.ports[portPath]; ok {
		port.Close()
		delete(sb.ports, portPath)
	}
	sb.mu.Unlock()

	// Reconectar este puerto específico después del delay
	go func() {
		for i := 0; i < sb.config.MaxReconnects; i++ {
			select {
			case <-sb.ctx.Done():
				return
			case <-time.After(sb.config.ReconnectDelay):
			}
			log.Printf("🔄 Reconectando %s (intento %d/%d)...", portPath, i+1, sb.config.MaxReconnects)
			if err := sb.connectPort(portPath); err == nil {
				log.Printf("✅ Reconexión exitosa en %s", portPath)
				return
			}
		}
		log.Printf("❌ No se pudo reconectar %s tras %d intentos", portPath, sb.config.MaxReconnects)
	}()
}

// tarifaChilquinta aplica la tarifa BT1 residencial de Chilquinta
// Cargo energía: $147.5 CLP/kWh + cargo fijo mensual $1.200 + IVA 19%
// calcularCostoChilquinta aplica tarifa BT-1A Chilquinta sin subsidio electrodependiente.
// Cargo energía: $239.0 CLP/kWh neto + cargo fijo mensual $2.344 neto + IVA 19%
// El cargo fijo se prorratea según el uptime del dispositivo en el período.
func calcularCostoChilquinta(energiaKwh float64, uptimeSegundos int64) float64 {
	const tarifaKwh = 239.0
	const cargoFijoMensual = 2344.0
	const iva = 1.19
	const segundosMes = 30.0 * 24.0 * 3600.0

	fraccionMes := float64(uptimeSegundos) / segundosMes
	if fraccionMes > 1.0 {
		fraccionMes = 1.0
	}

	return (energiaKwh*tarifaKwh + cargoFijoMensual*fraccionMes) * iva
}

func (sb *SerialBridge) processLine(line string, portPath string) {
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

	// Recalcular costo con tarifa real Chilquinta BT1 (ignora el costo que calcula el Arduino)
	data.Cost = calcularCostoChilquinta(data.Energy, data.Uptime)

	sb.devicesMu.Lock()
	device, exists := sb.devices[data.DeviceID]
	if !exists {
		device = &DeviceInfo{
			ID:       data.DeviceID,
			PortPath: portPath,
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
		// Si ya existe pero no tiene ClienteID, reintentar carga desde DB cada 10s
		needsRestore := device.ClienteID == "" && time.Since(device.LastClienteCheck) > 10*time.Second
		if needsRestore {
			device.LastClienteCheck = time.Now()
		}
		// Si el servicio debería estar activo pero el Arduino reporta 0 corriente, corregir
		needsServiceSync := device.ClienteID != "" && data.Current == 0 && data.Voltage == 0 &&
			!data.ServicioActivo && time.Since(device.LastClienteCheck) > 30*time.Second
		if needsServiceSync {
			device.LastClienteCheck = time.Now()
		}
		sb.devicesMu.Unlock()
		if needsRestore {
			go sb.restoreDeviceState(data.DeviceID)
		} else if needsServiceSync {
			go sb.syncServicioEstado(data.DeviceID)
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

	// Verificar umbral de costo para notificación SMS + Email
	go sb.checkCostThreshold(&data)
}

// checkCostThreshold verifica si el costo acumulado cruzó un umbral y envía SMS + Email
func (sb *SerialBridge) checkCostThreshold(data *ArduinoData) {
	const umbralCLP float64 = 1150 // Cada $15.000 CLP envía notificación

	sb.devicesMu.RLock()
	device := sb.devices[data.DeviceID]
	sb.devicesMu.RUnlock()
	if device == nil || device.ClienteID == "" {
		return
	}

	costoActual := data.Cost
	if costoActual < umbralCLP {
		return
	}

	// En qué múltiplo del umbral estamos: $15000, $30000, $45000...
	umbralActual := float64(int(costoActual/umbralCLP)) * umbralCLP

	// 1) Check rápido en memoria
	sb.costThresholdsMu.Lock()
	if umbralActual <= sb.costThresholds[data.DeviceID] {
		sb.costThresholdsMu.Unlock()
		return
	}
	sb.costThresholdsMu.Unlock()

	// 2) Check en MongoDB (sobrevive reinicios)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	thresholdCol := config.MongoDB.Collection("cost_thresholds")

	var doc bson.M
	_ = thresholdCol.FindOne(ctx, bson.M{"deviceId": data.DeviceID}).Decode(&doc)
	if doc != nil {
		if last, ok := doc["lastThreshold"].(float64); ok && umbralActual <= last {
			// Ya notificado — sincronizar memoria y salir
			sb.costThresholdsMu.Lock()
			sb.costThresholds[data.DeviceID] = last
			sb.costThresholdsMu.Unlock()
			return
		}
	}

	// 3) Nuevo umbral — persistir y notificar
	sb.costThresholdsMu.Lock()
	sb.costThresholds[data.DeviceID] = umbralActual
	sb.costThresholdsMu.Unlock()

	thresholdCol.UpdateOne(ctx,
		bson.M{"deviceId": data.DeviceID},
		bson.M{"$set": bson.M{"deviceId": data.DeviceID, "lastThreshold": umbralActual, "updatedAt": time.Now()}},
		options.Update().SetUpsert(true),
	)

	log.Printf("💰 Umbral de costo alcanzado: dispositivo %s → $%.0f CLP (energía: %.3f kWh, umbral: $%.0f)",
		data.DeviceID, costoActual, data.Energy, umbralActual)

	// Buscar datos del cliente en MongoDB
	clienteCol := config.MongoDB.Collection("clientes")
	clienteOID, err := primitive.ObjectIDFromHex(device.ClienteID)
	if err != nil {
		return
	}

	var cliente bson.M
	if err := clienteCol.FindOne(ctx, bson.M{"_id": clienteOID}).Decode(&cliente); err != nil {
		log.Printf("⚠️ No se encontró cliente %s para notificación de costo", device.ClienteID)
		return
	}

	nombreCliente, _ := cliente["nombre"].(string)
	telefono, _ := cliente["telefono"].(string)
	correo, _ := cliente["correo"].(string)

	// Enviar SMS si tiene teléfono
	if telefono != "" {
		go func() {
			smsService := sms.NewNoopService()
			mensaje := fmt.Sprintf(
				"⚡ Hola %s, tu consumo eléctrico ha alcanzado:\n"+
					"💡 %.2f kWh\n"+
					"💰 $%.0f CLP\n"+
					"Revisa tu consumo en la app.\n"+
					"- Electricautomaticchile",
				nombreCliente, data.Energy, costoActual,
			)
			if err := smsService.EnviarSMS(telefono, mensaje); err != nil {
				log.Printf("❌ Error enviando SMS de umbral a %s: %v", telefono, err)
			} else {
				log.Printf("✅ SMS de umbral enviado a %s ($%.0f CLP)", nombreCliente, costoActual)
			}
		}()
	}

	// Enviar Email con template de consumo
	if correo != "" {
		go func() {
			emailService := email.NewNoopService(config.AppConfig.EmailFrom)
			costoStr := fmt.Sprintf("$%.0f CLP", costoActual)
			energiaStr := fmt.Sprintf("%.2f kWh", data.Energy)

			subject := fmt.Sprintf("⚡ Tu consumo alcanzó %s - Electricautomaticchile", costoStr)

			htmlBody := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:Arial,sans-serif;color:#333">
<div style="max-width:600px;margin:0 auto">
<div style="background:linear-gradient(135deg,#f97316,#ea580c);color:white;padding:30px;text-align:center;border-radius:10px 10px 0 0">
<h1 style="margin:0">⚡ Electricautomaticchile</h1>
<p style="margin:5px 0 0;opacity:.9">Notificación de Consumo Eléctrico</p>
</div>
<div style="background:white;padding:30px;border-radius:0 0 10px 10px">
<p>Hola <strong>%s</strong>,</p>
<p>Tu consumo eléctrico ha alcanzado un nuevo umbral:</p>
<div style="background:#fff7ed;border:1px solid #fed7aa;border-radius:8px;padding:20px;text-align:center;margin:20px 0">
<div style="font-size:12px;color:#9a3412;text-transform:uppercase;letter-spacing:1px">Costo Acumulado</div>
<div style="font-size:36px;font-weight:bold;color:#ea580c">%s</div>
<div style="margin-top:10px;font-size:14px;color:#6b7280">%s &nbsp;·&nbsp; Dispositivo %s</div>
</div>
<p style="color:#6b7280;font-size:14px">Te recomendamos revisar tu consumo para optimizar tu gasto energético.</p>
<p><a href="https://electricautomaticchile.com/cliente" style="background:#f97316;color:white;padding:12px 30px;text-decoration:none;border-radius:8px;font-weight:bold;display:inline-block">Ver Mi Consumo</a></p>
</div>
<p style="text-align:center;color:#9ca3af;font-size:11px;margin-top:15px">Electricautomaticchile · Este es un correo automático, no responder.</p>
</div>
</body></html>`, nombreCliente, costoStr, energiaStr, data.DeviceID)

			textBody := fmt.Sprintf("Hola %s, tu consumo eléctrico ha alcanzado %s (%s). Revisa tu consumo en https://electricautomaticchile.com/cliente - Electricautomaticchile",
				nombreCliente, costoStr, energiaStr)

			if err := emailService.EnviarEmail([]string{correo}, subject, htmlBody, textBody); err != nil {
				log.Printf("❌ Error enviando email de consumo a %s: %v", correo, err)
			} else {
				log.Printf("✅ Email de consumo enviado a %s (%s)", correo, costoStr)
			}
		}()
	}
}

func (sb *SerialBridge) saveReading(data *ArduinoData) {
	// No guardar lecturas si el dispositivo no tiene cliente asignado
	sb.devicesMu.RLock()
	device := sb.devices[data.DeviceID]
	sb.devicesMu.RUnlock()
	if device == nil || device.ClienteID == "" {
		return
	}
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
			"corrienteMaxima": 1.43,
			"potenciaMaxima":  315,
			"tarifaKwh":       284,
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

	msg := eventbus.Message{
		Type:      eventbus.MessageTypeDeviceUpdate,
		Data:      wsData,
		Timestamp: time.Now(),
		ClienteID: device.ClienteID,
	}

	sent := false
	if device.ClienteID != "" {
		sb.publisher.PublishToCliente(device.ClienteID, msg)
		sent = true
	}
	if device.EmpresaID != "" {
		sb.publisher.PublishToEmpresa(device.EmpresaID, msg)
		sent = true
	}
	if !sent {
		// Dispositivo sin asignar — datos guardados en DB pero no enviados por WS
	}
}

func (sb *SerialBridge) SendCommand(command string) error {
	return sb.SendCommandToDevice("", command)
}

// SendCommandToDevice envía un comando al puerto del dispositivo específico
func (sb *SerialBridge) SendCommandToDevice(deviceID string, command string) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.ports) == 0 {
		return fmt.Errorf("no hay puertos conectados")
	}

	// Si se especifica deviceID, buscar su puerto
	if deviceID != "" {
		sb.devicesMu.RLock()
		device, ok := sb.devices[deviceID]
		sb.devicesMu.RUnlock()
		if ok && device.PortPath != "" {
			if port, exists := sb.ports[device.PortPath]; exists {
				log.Printf("📤 Enviando '%s' a %s (puerto %s)", command, deviceID, device.PortPath)
				_, err := port.Write([]byte(command + "\n"))
				return err
			}
		}
	}

	// Fallback: enviar a todos los puertos
	var lastErr error
	for portPath, port := range sb.ports {
		log.Printf("📤 Enviando '%s' a %s", command, portPath)
		if _, err := port.Write([]byte(command + "\n")); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (sb *SerialBridge) syncServicioEstado(deviceID string) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dispositivo bson.M
	if err := collection.FindOne(ctx, bson.M{"numeroDispositivo": deviceID}).Decode(&dispositivo); err != nil {
		return
	}

	estadoServicio := "activo"
	if es, ok := dispositivo["estadoServicio"].(string); ok && es != "" {
		estadoServicio = es
	}

	if estadoServicio == "cortado" {
		sb.SendCommandToDevice(deviceID, "DESACTIVAR_SERVICIO")
	} else {
		log.Printf("🔧 Sincronizando servicio de %s → ACTIVAR_SERVICIO", deviceID)
		sb.SendCommandToDevice(deviceID, "ACTIVAR_SERVICIO")
	}
}

func (sb *SerialBridge) restoreDeviceState(deviceID string) {
	collection := config.MongoDB.Collection("dispositivos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dispositivo bson.M
	if err := collection.FindOne(ctx, bson.M{"numeroDispositivo": deviceID}).Decode(&dispositivo); err != nil {
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

	// Restaurar estado del servicio (activo/cortado)
	estadoServicio := "activo"
	if es, ok := dispositivo["estadoServicio"].(string); ok && es != "" {
		estadoServicio = es
	}
	if estadoServicio == "cortado" {
		sb.SendCommandToDevice(deviceID, "DESACTIVAR_SERVICIO")
	} else {
		sb.SendCommandToDevice(deviceID, "ACTIVAR_SERVICIO")
	}

	if ultimaLectura, ok := dispositivo["ultimaLectura"].(bson.M); ok {
		if energia, ok := ultimaLectura["energia"].(float64); ok {
			costo := 0.0
			if c, ok := ultimaLectura["costo"].(float64); ok {
				costo = c
			}
			comando := fmt.Sprintf("RESTORE:%.3f:%.2f", energia, costo)
			sb.SendCommandToDevice(deviceID, comando)
		}
	}
}

func (sb *SerialBridge) IsConnected() bool {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return len(sb.ports) > 0
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
	for portPath, port := range sb.ports {
		port.Close()
		delete(sb.ports, portPath)
		log.Printf("✅ Puerto %s cerrado", portPath)
	}
}
