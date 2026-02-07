package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificacionEntity struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DestinatarioID  primitive.ObjectID `bson:"destinatarioId" json:"destinatarioId"`
	DispositivoID   primitive.ObjectID `bson:"dispositivoId,omitempty" json:"dispositivoId,omitempty"`
	Titulo          string             `bson:"titulo" json:"titulo"`
	Mensaje         string             `bson:"mensaje" json:"mensaje"`
	Tipo            string             `bson:"tipo" json:"tipo"`
	Severidad       string             `bson:"severidad,omitempty" json:"severidad,omitempty"`
	Leida           bool               `bson:"leida" json:"leida"`
	Resuelta        bool               `bson:"resuelta,omitempty" json:"resuelta,omitempty"`
	Importante      bool               `bson:"importante,omitempty" json:"importante,omitempty"`
	Resolucion      string             `bson:"resolucion,omitempty" json:"resolucion,omitempty"`
	FechaCreacion   time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaResolucion *time.Time         `bson:"fechaResolucion,omitempty" json:"fechaResolucion,omitempty"`
	Metadatos       map[string]interface{} `bson:"metadatos,omitempty" json:"metadatos,omitempty"`
}

func (NotificacionEntity) CollectionName() string {
	return "notificaciones"
}
