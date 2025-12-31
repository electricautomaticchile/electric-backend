package models

import "time"

type AlertaModel struct {
	ID               string                 `json:"_id" bson:"_id"`
	Tipo             string                 `json:"tipo"` // "error" | "advertencia" | "informacion" | "exito"
	Titulo           string                 `json:"titulo" binding:"required"`
	Mensaje          string                 `json:"mensaje" binding:"required"`
	Dispositivo      string                 `json:"dispositivo,omitempty"`
	EmpresaID        string                 `json:"empresaId" binding:"required"`
	Ubicacion        string                 `json:"ubicacion,omitempty"`
	Importante       bool                   `json:"importante"`
	Leida            bool                   `json:"leida"`
	Resuelta         bool                   `json:"resuelta"`
	AsignadoA        string                 `json:"asignadoA,omitempty"`
	FechaCreacion    time.Time              `json:"fechaCreacion"`
	FechaResolucion  *time.Time             `json:"fechaResolucion,omitempty"`
	AccionesTomadas  string                 `json:"accionesTomadas,omitempty"`
	Metadatos        map[string]interface{} `json:"metadatos,omitempty"`
}
