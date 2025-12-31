package arduino

import "time"

type ArduinoData struct {
	Type      string  `json:"type"`
	DeviceID  string  `json:"deviceId"`
	ClienteID string  `json:"clienteId"`
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
	Power     float64 `json:"activePower"`
	Energy    float64 `json:"energy"`
	Cost      float64 `json:"cost"`
	Uptime    int64   `json:"uptime"`
	LED1      bool    `json:"led1"`
	LED2      bool    `json:"led2"`
	Timestamp int64   `json:"timestamp"`
}

type DeviceInfo struct {
	ID          string
	ClienteID   string
	EmpresaID   string
	LastReading *ArduinoData
}

type SerialConfig struct {
	BaudRate        int
	DataBits        int
	StopBits        int
	Parity          string
	ReadTimeout     time.Duration
	ReconnectDelay  time.Duration
	MaxReconnects   int
	AutoRestore     bool
}
