package models

import "time"

type Reading struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
}

type LecturaDispositivoModel struct {
	Voltage     float64   `json:"voltage"`
	Current     float64   `json:"current"`
	ActivePower float64   `json:"activePower"`
	Energy      float64   `json:"energy"`
	Cost        float64   `json:"cost"`
	Timestamp   time.Time `json:"timestamp"`
}

type DispositivoModel struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"` // "LED" | "Sensor" | "Relay" | "Switch" | "Meter" | "Gateway"
	Status             bool                   `json:"status"`
	Location           string                 `json:"location,omitempty"`
	LastActivity       *time.Time             `json:"lastActivity,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	
	// Campos legacy para compatibilidad
	NumeroDispositivo  string                      `json:"numeroDispositivo,omitempty"`
	Nombre             string                      `json:"nombre,omitempty"`
	Tipo               string                      `json:"tipo,omitempty"`
	ClienteID          string                      `json:"clienteId,omitempty"`
	EmpresaID          string                      `json:"empresaId,omitempty"`
	Estado             string                      `json:"estado,omitempty"`
	UltimaLectura      *LecturaDispositivoModel    `json:"ultimaLectura,omitempty"`
	Configuracion      map[string]interface{}      `json:"configuracion,omitempty"`
	Activo             bool                        `json:"activo"`
	FechaCreacion      time.Time                   `json:"fechaCreacion,omitempty"`
	FechaActualizacion *time.Time                  `json:"fechaActualizacion,omitempty"`
}

type MeterDevice struct {
	DispositivoModel
	MeterType      string    `json:"meterType"` // "energy" | "water" | "gas"
	CurrentReading float64   `json:"currentReading,omitempty"`
	Unit           string    `json:"unit"`
	Readings       []Reading `json:"readings"`
}
