package services

import (
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/websocket"
	"time"
)

type WebSocketNotifierService struct {
	hub *websocket.Hub
}

func NewWebSocketNotifierService(hub *websocket.Hub) *WebSocketNotifierService {
	return &WebSocketNotifierService{
		hub: hub,
	}
}

func (s *WebSocketNotifierService) NotificarNuevaNotificacion(notificacion *entities.NotificacionEntity) {
	msg := websocket.Message{
		Type:      websocket.MessageTypeNotification,
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

	s.hub.BroadcastToCliente(notificacion.DestinatarioID.Hex(), msg)
	
	if notificacion.Tipo == "alerta" {
		alertMsg := websocket.Message{
			Type:      websocket.MessageTypeAlert,
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
		s.hub.BroadcastToEmpresa(notificacion.DestinatarioID.Hex(), alertMsg)
	}
}

func (s *WebSocketNotifierService) NotificarActualizacionDispositivo(dispositivo *entities.DispositivoEntity) {
	msg := websocket.Message{
		Type:      websocket.MessageTypeDeviceUpdate,
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

	s.hub.BroadcastToEmpresa(dispositivo.EmpresaID.Hex(), msg)
	
	if !dispositivo.ClienteID.IsZero() {
		s.hub.BroadcastToCliente(dispositivo.ClienteID.Hex(), msg)
	}
}

func (s *WebSocketNotifierService) NotificarConsumo(empresaID string, clienteID string, data map[string]interface{}) {
	msg := websocket.Message{
		Type:      websocket.MessageTypeConsumption,
		Timestamp: time.Now(),
		EmpresaID: empresaID,
		ClienteID: clienteID,
		Data:      data,
	}

	if empresaID != "" {
		s.hub.BroadcastToEmpresa(empresaID, msg)
	}
	
	if clienteID != "" {
		s.hub.BroadcastToCliente(clienteID, msg)
	}
}
