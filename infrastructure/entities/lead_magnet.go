package entities

import (
"time"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadMagnetEntity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         string             `bson:"email" json:"email"`
	Nombre        string             `bson:"nombre" json:"nombre"`
	FechaCreacion time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
}

func (LeadMagnetEntity) CollectionName() string {
	return "leadmagnets"
}
