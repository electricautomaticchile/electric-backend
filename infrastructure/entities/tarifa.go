package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TramosEstabilizacion struct {
	Hasta350Kwh   float64 `bson:"hasta350Kwh" json:"hasta350Kwh"`
	Entre350Y500  float64 `bson:"entre350Y500" json:"entre350Y500"`
	Mayor500Kwh   float64 `bson:"mayor500Kwh" json:"mayor500Kwh"`
}

type TarifaEntity struct {
	ID                    primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Distribuidora         string               `bson:"distribuidora" json:"distribuidora"`
	TipoTarifa            string               `bson:"tipoTarifa" json:"tipoTarifa"`
	Comuna                string               `bson:"comuna" json:"comuna"`
	RedTipo               string               `bson:"redTipo" json:"redTipo"`
	VigenciaDesde         time.Time            `bson:"vigenciaDesde" json:"vigenciaDesde"`
	VigenciaHasta         time.Time            `bson:"vigenciaHasta" json:"vigenciaHasta"`
	CargoEnergia          float64              `bson:"cargoEnergia" json:"cargoEnergia"`
	CargoTransmision      float64              `bson:"cargoTransmision" json:"cargoTransmision"`
	CargoServicioPublico  float64              `bson:"cargoServicioPublico" json:"cargoServicioPublico"`
	PrecioKwhBase         float64              `bson:"precioKwhBase" json:"precioKwhBase"`
	PeajeDistribucion     float64              `bson:"peajeDistribucion" json:"peajeDistribucion"`
	TramosEstabilizacion  TramosEstabilizacion `bson:"tramosEstabilizacion" json:"tramosEstabilizacion"`
	Activa                bool                 `bson:"activa" json:"activa"`
	FechaCreacion         time.Time            `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaActualizacion    time.Time            `bson:"fechaActualizacion" json:"fechaActualizacion"`
}

func (TarifaEntity) CollectionName() string {
	return "tarifas"
}
