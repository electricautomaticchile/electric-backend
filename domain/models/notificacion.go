package models

import "time"

type NotificacionModel struct {
	ID             string    `json:"id"`
	DestinatarioID string    `json:"destinatarioId"`
	Titulo         string    `json:"titulo"`
	Mensaje        string    `json:"mensaje"`
	Tipo           string    `json:"tipo"`
	Leida          bool      `json:"leida"`
	FechaCreacion  time.Time `json:"fechaCreacion"`
}
