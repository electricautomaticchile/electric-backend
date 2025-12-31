package entities

import (
"time"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificacionEntity struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DestinatarioID primitive.ObjectID `bson:"destinatarioId" json:"destinatarioId"`
	Titulo         string             `bson:"titulo" json:"titulo"`
	Mensaje        string             `bson:"mensaje" json:"mensaje"`
	Tipo           string             `bson:"tipo" json:"tipo"`
	Leida          bool               `bson:"leida" json:"leida"`
	FechaCreacion  time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
}

func (NotificacionEntity) CollectionName() string {
	return "notificaciones"
}
