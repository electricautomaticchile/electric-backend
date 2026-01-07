package models

import "time"

type TramosEstabilizacion struct {
	Hasta350Kwh   float64 `json:"hasta350Kwh"`
	Entre350Y500  float64 `json:"entre350Y500"`
	Mayor500Kwh   float64 `json:"mayor500Kwh"`
}

type TarifaModel struct {
	ID                    string               `json:"_id" bson:"_id"`
	Distribuidora         string               `json:"distribuidora" binding:"required"`
	TipoTarifa            string               `json:"tipoTarifa" binding:"required"`
	Comuna                string               `json:"comuna" binding:"required"`
	RedTipo               string               `json:"redTipo" binding:"required"`
	VigenciaDesde         time.Time            `json:"vigenciaDesde" binding:"required"`
	VigenciaHasta         time.Time            `json:"vigenciaHasta" binding:"required"`
	CargoEnergia          float64              `json:"cargoEnergia" binding:"required"`
	CargoTransmision      float64              `json:"cargoTransmision" binding:"required"`
	CargoServicioPublico  float64              `json:"cargoServicioPublico" binding:"required"`
	PrecioKwhBase         float64              `json:"precioKwhBase" binding:"required"`
	PeajeDistribucion     float64              `json:"peajeDistribucion"`
	TramosEstabilizacion  TramosEstabilizacion `json:"tramosEstabilizacion" binding:"required"`
	Activa                bool                 `json:"activa"`
	FechaCreacion         time.Time            `json:"fechaCreacion"`
	FechaActualizacion    time.Time            `json:"fechaActualizacion"`
}

type CalculoConsumoRequest struct {
	KwhConsumidos float64 `json:"kwhConsumidos" binding:"required"`
	TarifaID      string  `json:"tarifaId" binding:"required"`
}

type CalculoConsumoResponse struct {
	KwhConsumidos         float64 `json:"kwhConsumidos"`
	PrecioKwhBase         float64 `json:"precioKwhBase"`
	CargoEstabilizacion   float64 `json:"cargoEstabilizacion"`
	MontoBase             float64 `json:"montoBase"`
	MontoEstabilizacion   float64 `json:"montoEstabilizacion"`
	MontoTotal            float64 `json:"montoTotal"`
	TramoEstabilizacion   string  `json:"tramoEstabilizacion"`
}
