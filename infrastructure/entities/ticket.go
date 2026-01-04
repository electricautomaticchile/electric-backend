package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TicketEntity struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NumeroTicket   string             `bson:"numeroTicket" json:"numeroTicket"`
	Titulo         string             `bson:"titulo" json:"titulo"`
	Descripcion    string             `bson:"descripcion" json:"descripcion"`
	Estado         string             `bson:"estado" json:"estado"`
	Prioridad      string             `bson:"prioridad" json:"prioridad"`
	Categoria      string             `bson:"categoria" json:"categoria"`
	ClienteID      primitive.ObjectID `bson:"clienteId,omitempty" json:"clienteId,omitempty"`
	EmpresaID      primitive.ObjectID `bson:"empresaId,omitempty" json:"empresaId,omitempty"`
	Respuestas     []RespuestaTicket  `bson:"respuestas" json:"respuestas"`
	FechaCreacion  time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
}

type RespuestaTicket struct {
	Mensaje       string    `bson:"mensaje" json:"mensaje"`
	UsuarioID     string    `bson:"usuarioId" json:"usuarioId"`
	FechaCreacion time.Time `bson:"fechaCreacion" json:"fechaCreacion"`
}

func (TicketEntity) CollectionName() string {
	return "tickets"
}
