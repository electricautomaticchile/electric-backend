package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FCMTokenEntity representa un token de Firebase Cloud Messaging asociado a un
// usuario. Un usuario puede tener múltiples tokens (multi-dispositivo).
type FCMTokenEntity struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Token      string             `bson:"token" json:"token"`
	Plataforma string             `bson:"plataforma" json:"plataforma"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func (FCMTokenEntity) CollectionName() string {
	return "fcm_tokens"
}
