package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UmbralAlertaEntity representa los umbrales de alerta configurables por
// empresa para el monitoreo de dispositivos. Existe como máximo un documento
// por empresa (identificado por empresaId). Si una empresa no tiene documento,
// el servicio aplica los umbrales por defecto.
type UmbralAlertaEntity struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EmpresaID    string             `bson:"empresaId" json:"empresaId"`
	VoltajeMin   float64            `bson:"voltajeMin" json:"voltajeMin"`
	VoltajeMax   float64            `bson:"voltajeMax" json:"voltajeMax"`
	CorrienteMax float64            `bson:"corrienteMax" json:"corrienteMax"`
	ConsumoMax   float64            `bson:"consumoMax" json:"consumoMax"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func (UmbralAlertaEntity) CollectionName() string {
	return "umbrales_alerta"
}
