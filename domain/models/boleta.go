package models

import "time"

type BoletaModel struct {
	ID            string     `json:"id"`
	ClienteID     string     `json:"clienteId"`
	Monto         float64    `json:"monto"`
	Periodo       string     `json:"periodo"`
	Estado        string     `json:"estado"`
	FechaCreacion time.Time  `json:"fechaCreacion"`
	FechaPago     *time.Time `json:"fechaPago,omitempty"`
}
