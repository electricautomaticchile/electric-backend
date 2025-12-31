package websocket

import "time"

type MessageType string

const (
	MessageTypeAlert        MessageType = "alert"
	MessageTypeNotification MessageType = "notification"
	MessageTypeDeviceUpdate MessageType = "device_update"
	MessageTypeConsumption  MessageType = "consumption"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

type Message struct {
	Type      MessageType            `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	EmpresaID string                 `json:"empresaId,omitempty"`
	ClienteID string                 `json:"clienteId,omitempty"`
}

type ClientMessage struct {
	Type   string                 `json:"type"`
	Action string                 `json:"action,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}
