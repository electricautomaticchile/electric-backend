package entities

import (
"time"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertaEntity struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EmpresaID      primitive.ObjectID `bson:"empresaId" json:"empresaId"`
	DispositivoID  primitive.ObjectID `bson:"dispositivoId,omitempty" json:"dispositivoId,omitempty"`
	Tipo           string             `bson:"tipo" json:"tipo"`
	Severidad      string             `bson:"severidad" json:"severidad"`
	Mensaje        string             `bson:"mensaje" json:"mensaje"`
	Resuelta       bool               `bson:"resuelta" json:"resuelta"`
	Resolucion     string             `bson:"resolucion,omitempty" json:"resolucion,omitempty"`
	FechaCreacion  time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaResolucion *time.Time        `bson:"fechaResolucion,omitempty" json:"fechaResolucion,omitempty"`
}

func (AlertaEntity) CollectionName() string {
	return "alertas"
}
