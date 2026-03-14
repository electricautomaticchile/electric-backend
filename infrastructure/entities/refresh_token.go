package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshTokenEntity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Token       string             `bson:"token" json:"token"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserType    string             `bson:"userType" json:"userType"`
	Fingerprint string             `bson:"fingerprint,omitempty" json:"fingerprint,omitempty"` // MED-08
	ExpiresAt   time.Time          `bson:"expiresAt" json:"expiresAt"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	Revoked     bool               `bson:"revoked" json:"revoked"`
	RevokedAt   *time.Time         `bson:"revokedAt,omitempty" json:"revokedAt,omitempty"`
}

func (RefreshTokenEntity) CollectionName() string {
	return "refresh_tokens"
}
