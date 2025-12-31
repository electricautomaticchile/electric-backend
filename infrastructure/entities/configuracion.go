package entities

import (
"time"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConfiguracionEntity struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Clave              string             `bson:"clave" json:"clave"`
	Valor              interface{}        `bson:"valor" json:"valor"`
	Categoria          string             `bson:"categoria" json:"categoria"`
	FechaActualizacion time.Time          `bson:"fechaActualizacion" json:"fechaActualizacion"`
}

func (ConfiguracionEntity) CollectionName() string {
	return "configuraciones"
}
