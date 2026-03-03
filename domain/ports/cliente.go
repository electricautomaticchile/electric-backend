package ports

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/types"
)

type PortCliente interface {
	FindAll(ctx context.Context, empresaID string) ([]*models.ClienteModel, error)
	FindAllPaginated(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) ([]*models.ClienteModel, int64, error)
	FindByID(ctx context.Context, id string) (*models.ClienteModel, error)
	FindByNumeroCliente(ctx context.Context, numeroCliente string) (*models.ClienteModel, error)
	FindByRut(ctx context.Context, rut string) (*models.ClienteModel, error)
	FindByCorreo(ctx context.Context, correo string) (*models.ClienteModel, error)
	Create(ctx context.Context, cliente *models.ClienteModel) error
	Update(ctx context.Context, id string, cliente *models.ClienteModel) error
	UpdateUltimoAcceso(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	Delete(ctx context.Context, id string) error
}
