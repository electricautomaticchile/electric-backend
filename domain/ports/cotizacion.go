package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortCotizacion interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.CotizacionModel, int64, error)
	FindByID(ctx context.Context, id string) (*models.CotizacionModel, error)
	FindByNumero(ctx context.Context, numero string) (*models.CotizacionModel, error)
	Create(ctx context.Context, cotizacion *models.CotizacionModel) error
	Update(ctx context.Context, id string, cotizacion *models.CotizacionModel) error
	UpdateEstado(ctx context.Context, id string, estado string) error
	Delete(ctx context.Context, id string) error
}
