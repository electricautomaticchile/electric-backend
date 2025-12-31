package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortPower interface {
	FindByUsuarioID(ctx context.Context, usuarioID string) ([]string, error)
	Create(ctx context.Context, power *entities.PowerEntity) error
	Delete(ctx context.Context, usuarioID string, power string) error
	DeleteAllByUsuarioID(ctx context.Context, usuarioID string) error
}
