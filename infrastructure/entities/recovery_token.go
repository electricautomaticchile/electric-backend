package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecoveryTokenEntity struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UsuarioID      primitive.ObjectID `bson:"usuarioId" json:"usuarioId"`
	Token          string             `bson:"token" json:"token"`
	Usado          bool               `bson:"usado" json:"usado"`
	FechaCreacion  time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaExpiracion time.Time         `bson:"fechaExpiracion" json:"fechaExpiracion"`
}

func (RecoveryTokenEntity) CollectionName() string {
	return "recoverytokens"
}
