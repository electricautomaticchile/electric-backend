package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortAlerta interface {
	FindAll(ctx context.Context) ([]*models.AlertaModel, error)
	FindActivas(ctx context.Context) ([]*models.AlertaModel, error)
	FindByEmpresa(ctx context.Context, empresaID string) ([]*models.AlertaModel, error)
	FindByID(ctx context.Context, id string) (*models.AlertaModel, error)
	Create(ctx context.Context, alerta *models.AlertaModel) error
	Resolver(ctx context.Context, id string, resolucion string) error
	Delete(ctx context.Context, id string) error
}
