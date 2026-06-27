package services

import (
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/eventbus"
	"time"
)

// WebSocketNotifierService publica eventos en tiempo real en Redis Pub/Sub.
// El broadcast a los clientes WebSocket lo realiza el servicio independiente
// websocket-electric, que se suscribe a estos eventos. La API ya no mantiene
// conexiones WebSocket.
type WebSocketNotifierService struct {
	publisher *eventbus.Publisher
}

func NewWebSocketNotifierService(publisher *eventbus.Publisher) *WebSocketNotifierService {
	return &WebSocketNotifierService{
		publisher: publisher,
	}
}

func (s *WebSocketNotifierService) NotificarNuevaNotificacion(notificacion *entities.NotificacionEntity) {
	if s.publisher == nil {
		return
	}

	msg := eventbus.Message{
		Type:      eventbus.MessageTypeNotification,
		Timestamp: time.Now(),
		ClienteID: notificacion.DestinatarioID.Hex(),
		Data: map[string]interface{}{
			"id":      notificacion.ID.Hex(),
			"tipo":    notificacion.Tipo,
			"titulo":  notificacion.Titulo,
			"mensaje": notificacion.Mensaje,
			"leida":   notificacion.Leida,
		},
	}

	s.publisher.PublishToCliente(notificacion.DestinatarioID.Hex(), msg)

	if notificacion.Tipo == "alerta" {
		alertMsg := eventbus.Message{
			Type:      eventbus.MessageTypeAlert,
			Timestamp: time.Now(),
			EmpresaID: notificacion.DestinatarioID.Hex(),
			Data: map[string]interface{}{
				"id":          notificacion.ID.Hex(),
				"tipo":        notificacion.Tipo,
				"mensaje":     notificacion.Mensaje,
				"severidad":   notificacion.Severidad,
				"dispositivo": notificacion.DispositivoID.Hex(),
				"resuelta":    notificacion.Resuelta,
			},
		}
		s.publisher.PublishToEmpresa(notificacion.DestinatarioID.Hex(), alertMsg)
	}
}

func (s *WebSocketNotifierService) NotificarActualizacionDispositivo(dispositivo *entities.DispositivoEntity) {
	if s.publisher == nil {
		return
	}

	msg := eventbus.Message{
		Type:      eventbus.MessageTypeDeviceUpdate,
		Timestamp: time.Now(),
		EmpresaID: dispositivo.EmpresaID.Hex(),
		Data: map[string]interface{}{
			"id":                dispositivo.ID.Hex(),
			"numeroDispositivo": dispositivo.NumeroDispositivo,
			"nombre":            dispositivo.Nombre,
			"estado":            dispositivo.Estado,
			"clienteId":         dispositivo.ClienteID.Hex(),
		},
	}

	if dispositivo.UltimaLectura != nil {
		msg.Data["ultimaLectura"] = map[string]interface{}{
			"voltage":     dispositivo.UltimaLectura.Voltage,
			"current":     dispositivo.UltimaLectura.Current,
			"activePower": dispositivo.UltimaLectura.ActivePower,
			"energy":      dispositivo.UltimaLectura.Energy,
			"cost":        dispositivo.UltimaLectura.Cost,
			"timestamp":   dispositivo.UltimaLectura.Timestamp,
		}
	}

	s.publisher.PublishToEmpresa(dispositivo.EmpresaID.Hex(), msg)

	if !dispositivo.ClienteID.IsZero() {
		s.publisher.PublishToCliente(dispositivo.ClienteID.Hex(), msg)
	}
}

func (s *WebSocketNotifierService) NotificarConsumo(empresaID string, clienteID string, data map[string]interface{}) {
	if s.publisher == nil {
		return
	}

	msg := eventbus.Message{
		Type:      eventbus.MessageTypeConsumption,
		Timestamp: time.Now(),
		EmpresaID: empresaID,
		ClienteID: clienteID,
		Data:      data,
	}

	if empresaID != "" {
		s.publisher.PublishToEmpresa(empresaID, msg)
	}

	if clienteID != "" {
		s.publisher.PublishToCliente(clienteID, msg)
	}
}
