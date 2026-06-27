package eventbus

import "time"

// MessageType identifica el tipo de mensaje en tiempo real entregado al cliente.
type MessageType string

const (
	MessageTypeAlert        MessageType = "alert"
	MessageTypeNotification MessageType = "notification"
	MessageTypeDeviceUpdate MessageType = "device_update"
	MessageTypeConsumption  MessageType = "consumption"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// Message es el payload en tiempo real que el WS Hub entrega a los clientes.
// El contrato debe coincidir con el del servicio websocket-electric.
type Message struct {
	Type      MessageType            `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	EmpresaID string                 `json:"empresaId,omitempty"`
	ClienteID string                 `json:"clienteId,omitempty"`
}

// Event es el sobre publicado en Redis Pub/Sub. El WS Hub se suscribe, lo
// deserializa y enruta el Message según Scope/TargetID.
type Event struct {
	Scope    string  `json:"scope"`
	TargetID string  `json:"targetId,omitempty"`
	Message  Message `json:"message"`
}

// Scopes de destino.
const (
	ScopeCliente = "cliente"
	ScopeEmpresa = "empresa"
	ScopeAll     = "all"
)

// Canales de Redis Pub/Sub (deben coincidir con los del WS Hub).
const (
	ChannelEvents    = "ws:events"
	ChannelBroadcast = "ws:broadcast"
)
