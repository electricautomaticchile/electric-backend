package dto

import "time"

// DtoDispositivo es la respuesta pública de un dispositivo (sin iotToken).
type DtoDispositivo struct {
	ID                string          `json:"id"`
	NumeroDispositivo string          `json:"numeroDispositivo"`
	Nombre            string          `json:"nombre"`
	Tipo              string          `json:"tipo"`
	Estado            string          `json:"estado"`
	Activo            bool            `json:"activo"`
	UltimaLectura     *DtoLectura    `json:"ultimaLectura,omitempty"`
	FechaCreacion     time.Time       `json:"fechaCreacion"`
}

// DtoLectura es la respuesta pública de una lectura de dispositivo.
type DtoLectura struct {
	Voltage    float64   `json:"voltage"`
	Current    float64   `json:"current"`
	Power      float64   `json:"activePower"`
	Energy     float64   `json:"energy"`
	Cost       float64   `json:"cost"`
	Timestamp  time.Time `json:"timestamp"`
}

// DtoIoTStatus es la respuesta del endpoint /api/v1/iot/status.
type DtoIoTStatus struct {
	ID                string     `json:"id"`
	NumeroDispositivo string     `json:"numeroDispositivo"`
	Nombre            string     `json:"nombre"`
	Estado            string     `json:"estado"`
	Online            bool       `json:"online"`
	UltimaLectura     *time.Time `json:"ultimaLectura,omitempty"`
	SenalGSM          string     `json:"senalGSM"`
}
