package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortTarifa interface {
	Create(ctx context.Context, tarifa *models.TarifaModel) error
	FindByID(ctx context.Context, id string) (*models.TarifaModel, error)
	FindActiva(ctx context.Context, comuna, tipoTarifa string) (*models.TarifaModel, error)
	FindAll(ctx context.Context) ([]*models.TarifaModel, error)
	Update(ctx context.Context, id string, tarifa *models.TarifaModel) error
	Delete(ctx context.Context, id string) error
}
