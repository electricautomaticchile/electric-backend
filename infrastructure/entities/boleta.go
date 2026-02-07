package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoletaEntity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ClienteID     primitive.ObjectID `bson:"clienteId" json:"clienteId"`
	Monto         float64            `bson:"monto" json:"monto"`
	MontoTotal    float64            `bson:"montoTotal" json:"montoTotal"`
	Periodo       string             `bson:"periodo" json:"periodo"`
	Estado        string             `bson:"estado" json:"estado"`
	FechaCreacion time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaPago     *time.Time         `bson:"fechaPago,omitempty" json:"fechaPago,omitempty"`
}

func (BoletaEntity) CollectionName() string {
	return "boletas"
}
