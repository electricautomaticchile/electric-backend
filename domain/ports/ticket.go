package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortTicket interface {
	FindAll(ctx context.Context) ([]*entities.TicketEntity, error)
	FindByID(ctx context.Context, id string) (*entities.TicketEntity, error)
	FindByCliente(ctx context.Context, clienteID string) ([]*entities.TicketEntity, error)
	FindByEmpresa(ctx context.Context, empresaID string) ([]*entities.TicketEntity, error)
	Create(ctx context.Context, ticket *entities.TicketEntity) error
	AgregarRespuesta(ctx context.Context, id string, respuesta *entities.RespuestaTicket) error
	ActualizarEstado(ctx context.Context, id string, estado string) error
	Delete(ctx context.Context, id string) error
}
