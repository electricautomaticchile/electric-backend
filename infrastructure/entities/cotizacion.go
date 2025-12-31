package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CotizacionEntity struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Numero             string             `bson:"numero,omitempty" json:"numero,omitempty"`
	Nombre             string             `bson:"nombre" json:"nombre"`
	Email              string             `bson:"email" json:"email"`
	Empresa            string             `bson:"empresa,omitempty" json:"empresa,omitempty"`
	Telefono           string             `bson:"telefono,omitempty" json:"telefono,omitempty"`
	Servicio           string             `bson:"servicio" json:"servicio"`
	Plazo              string             `bson:"plazo,omitempty" json:"plazo,omitempty"`
	Mensaje            string             `bson:"mensaje" json:"mensaje"`
	Estado             string             `bson:"estado" json:"estado"`
	Prioridad          string             `bson:"prioridad,omitempty" json:"prioridad,omitempty"`
	FechaCreacion      interface{}        `bson:"fechaCreacion,omitempty" json:"fechaCreacion,omitempty"`
	FechaActualizacion interface{}        `bson:"fechaActualizacion,omitempty" json:"fechaActualizacion,omitempty"`
}

func (CotizacionEntity) CollectionName() string {
	return "cotizaciones"
}
