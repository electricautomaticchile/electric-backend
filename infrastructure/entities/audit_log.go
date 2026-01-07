package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLogEntity struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID        string                 `bson:"userId" json:"userId"`
	UserType      string                 `bson:"userType" json:"userType"`
	UserEmail     string                 `bson:"userEmail" json:"userEmail"`
	EmpresaID     string                 `bson:"empresaId,omitempty" json:"empresaId,omitempty"`
	Action        string                 `bson:"action" json:"action"`
	Resource      string                 `bson:"resource" json:"resource"`
	ResourceID    string                 `bson:"resourceId,omitempty" json:"resourceId,omitempty"`
	Method        string                 `bson:"method" json:"method"`
	Endpoint      string                 `bson:"endpoint" json:"endpoint"`
	IPAddress     string                 `bson:"ipAddress" json:"ipAddress"`
	UserAgent     string                 `bson:"userAgent" json:"userAgent"`
	StatusCode    int                    `bson:"statusCode" json:"statusCode"`
	Success       bool                   `bson:"success" json:"success"`
	ErrorMessage  string                 `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	RequestBody   map[string]interface{} `bson:"requestBody,omitempty" json:"requestBody,omitempty"`
	ResponseBody  map[string]interface{} `bson:"responseBody,omitempty" json:"responseBody,omitempty"`
	Duration      int64                  `bson:"duration" json:"duration"`
	Timestamp     time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata      map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

func (AuditLogEntity) CollectionName() string {
	return "audit_logs"
}
