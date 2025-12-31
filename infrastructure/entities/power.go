package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PowerEntity struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UsuarioID          primitive.ObjectID `bson:"usuarioId" json:"usuarioId"`
	Power              string             `bson:"power" json:"power"`
	FechaCreacion      time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaActualizacion *time.Time         `bson:"fechaActualizacion,omitempty" json:"fechaActualizacion,omitempty"`
}

func (PowerEntity) CollectionName() string {
	return "powers"
}
