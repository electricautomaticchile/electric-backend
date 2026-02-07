package models

import "time"

type NotificacionModel struct {
	ID              string                 `json:"id"`
	DestinatarioID  string                 `json:"destinatarioId"`
	DispositivoID   string                 `json:"dispositivoId,omitempty"`
	Titulo          string                 `json:"titulo"`
	Mensaje         string                 `json:"mensaje"`
	Tipo            string                 `json:"tipo"`
	Severidad       string                 `json:"severidad,omitempty"`
	Leida           bool                   `json:"leida"`
	Resuelta        bool                   `json:"resuelta,omitempty"`
	Importante      bool                   `json:"importante,omitempty"`
	Resolucion      string                 `json:"resolucion,omitempty"`
	FechaCreacion   time.Time              `json:"fechaCreacion"`
	FechaResolucion *time.Time             `json:"fechaResolucion,omitempty"`
	Metadatos       map[string]interface{} `json:"metadatos,omitempty"`
}
