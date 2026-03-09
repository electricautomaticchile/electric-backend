package arduino

import "time"

type ArduinoData struct {
	Type           string  `json:"type"`
	DeviceID       string  `json:"deviceId"`
	ClienteID      string  `json:"clienteId"`
	Voltage        float64 `json:"voltage"`
	Current        float64 `json:"current"`
	Power          float64 `json:"power"`
	Energy         float64 `json:"energy"`
	Cost           float64 `json:"cost"`
	Uptime         int64   `json:"uptime"`
	ServicioActivo bool    `json:"servicioActivo"`
	Timestamp      int64   `json:"timestamp"`
}

type DeviceInfo struct {
	ID               string
	ClienteID        string
	EmpresaID        string
	LastReading      *ArduinoData
	PortPath         string
	LastClienteCheck time.Time
}

type SerialConfig struct {
	BaudRate       int
	DataBits       int
	StopBits       int
	Parity         string
	ReadTimeout    time.Duration
	ReconnectDelay time.Duration
	MaxReconnects  int
	AutoRestore    bool
}
