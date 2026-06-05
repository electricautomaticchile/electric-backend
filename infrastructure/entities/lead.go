package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadEntity struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Type         string                 `bson:"type" json:"type"`
	Status       string                 `bson:"status" json:"status"`
	Name         string                 `bson:"name" json:"name"`
	Email        string                 `bson:"email" json:"email"`
	Organization string                 `bson:"organization,omitempty" json:"organization,omitempty"`
	Message      string                 `bson:"message,omitempty" json:"message,omitempty"`
	Extra        map[string]interface{} `bson:"extra,omitempty" json:"extra,omitempty"`
	IPAddress    string                 `bson:"ipAddress,omitempty" json:"-"`
	UserAgent    string                 `bson:"userAgent,omitempty" json:"-"`
	CreatedAt    time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time              `bson:"updatedAt" json:"updatedAt"`
}

func (LeadEntity) CollectionName() string {
	return "leads"
}
