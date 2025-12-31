package models

import "time"

type TicketModel struct {
	ID            string              `json:"id"`
	NumeroTicket  string              `json:"numeroTicket"`
	Titulo        string              `json:"titulo"`
	Descripcion   string              `json:"descripcion"`
	Estado        string              `json:"estado"`
	Prioridad     string              `json:"prioridad"`
	ClienteID     string              `json:"clienteId,omitempty"`
	Respuestas    []RespuestaTicketModel `json:"respuestas"`
	FechaCreacion time.Time           `json:"fechaCreacion"`
}

type RespuestaTicketModel struct {
	Mensaje       string    `json:"mensaje"`
	UsuarioID     string    `json:"usuarioId"`
	FechaCreacion time.Time `json:"fechaCreacion"`
}
