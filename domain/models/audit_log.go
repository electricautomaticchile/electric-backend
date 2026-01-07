package models

import "time"

type AuditLogModel struct {
	ID            string                 `json:"_id" bson:"_id"`
	UserID        string                 `json:"userId"`
	UserType      string                 `json:"userType"`
	UserEmail     string                 `json:"userEmail"`
	EmpresaID     string                 `json:"empresaId,omitempty"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	ResourceID    string                 `json:"resourceId,omitempty"`
	Method        string                 `json:"method"`
	Endpoint      string                 `json:"endpoint"`
	IPAddress     string                 `json:"ipAddress"`
	UserAgent     string                 `json:"userAgent"`
	StatusCode    int                    `json:"statusCode"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	RequestBody   map[string]interface{} `json:"requestBody,omitempty"`
	ResponseBody  map[string]interface{} `json:"responseBody,omitempty"`
	Duration      int64                  `json:"duration"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
