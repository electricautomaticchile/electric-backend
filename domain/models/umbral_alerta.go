package models

import "time"

// UmbralAlertaModel es la representación de API de los umbrales de alerta de una
// empresa. Cuando la empresa no tiene configuración propia, se devuelven los
// valores por defecto y EsDefault=true.
type UmbralAlertaModel struct {
	EmpresaID    string    `json:"empresaId"`
	VoltajeMin   float64   `json:"voltajeMin"`
	VoltajeMax   float64   `json:"voltajeMax"`
	CorrienteMax float64   `json:"corrienteMax"`
	ConsumoMax   float64   `json:"consumoMax"`
	EsDefault    bool      `json:"esDefault"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
}
