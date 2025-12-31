package ports

import (
"context"
"electric-backend/infrastructure/entities"
)

type PortBoleta interface {
	FindByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error)
	FindByID(ctx context.Context, id string) (*entities.BoletaEntity, error)
	Create(ctx context.Context, boleta *entities.BoletaEntity) error
}
