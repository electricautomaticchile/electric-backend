package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortNotificacion interface {
	FindByDestinatario(ctx context.Context, destinatarioID string) ([]*entities.NotificacionEntity, error)
	FindByEmpresa(ctx context.Context, empresaID string) ([]*entities.NotificacionEntity, error)
	FindActivas(ctx context.Context, empresaID string) ([]*entities.NotificacionEntity, error)
	FindByID(ctx context.Context, id string) (*entities.NotificacionEntity, error)
	Create(ctx context.Context, notificacion *entities.NotificacionEntity) error
	MarcarLeida(ctx context.Context, id string) error
	MarcarTodasLeidas(ctx context.Context, destinatarioID string) error
	Resolver(ctx context.Context, id string, resolucion string) error
	Delete(ctx context.Context, id string) error
}
