package models

import "time"

type BoletaModel struct {
	ID               string     `json:"id"`
	ClienteID        string     `json:"clienteId"`
	EmpresaID        string     `json:"empresaId,omitempty"`
	DispositivoID    string     `json:"dispositivoId,omitempty"`
	Monto            float64    `json:"monto"`
	Periodo          string     `json:"periodo"`
	Mes              int        `json:"mes"`
	Anio             int        `json:"anio"`
	ConsumoKwh       float64    `json:"consumoKwh"`
	Estado           string     `json:"estado"`
	FechaCreacion    time.Time  `json:"fechaCreacion"`
	FechaVencimiento *time.Time `json:"fechaVencimiento,omitempty"`
	FechaPago        *time.Time `json:"fechaPago,omitempty"`
	MotivoCorte      string     `json:"motivoCorte,omitempty"`
}

type DeudaResumenModel struct {
	BoletasPendientes  int        `json:"boletasPendientes"`
	BoletasVencidas    int        `json:"boletasVencidas"`
	MontoTotal         float64    `json:"montoTotal"`
	MontoVencido       float64    `json:"montoVencido"`
	ProximoVencimiento *time.Time `json:"proximoVencimiento,omitempty"`
	NivelAlerta        string     `json:"nivelAlerta"`
}

type ConfirmarPagoResponse struct {
	BoletaID         string    `json:"boletaId"`
	Estado           string    `json:"estado"`
	FechaPago        time.Time `json:"fechaPago"`
	ServicioRepuesto bool      `json:"servicioRepuesto"`
	Mensaje          string    `json:"mensaje"`
}
