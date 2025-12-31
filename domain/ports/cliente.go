package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortCliente interface {
	FindAll(ctx context.Context, empresaID string) ([]*models.ClienteModel, error)
	FindByID(ctx context.Context, id string) (*models.ClienteModel, error)
	FindByNumero(ctx context.Context, numeroCliente string) (*models.ClienteModel, error)
	FindByNumeroCliente(ctx context.Context, numeroCliente string) (*models.ClienteModel, error)
	Create(ctx context.Context, cliente *models.ClienteModel) error
	Update(ctx context.Context, id string, cliente *models.ClienteModel) error
	UpdateUltimoAcceso(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	Delete(ctx context.Context, id string) error
}
