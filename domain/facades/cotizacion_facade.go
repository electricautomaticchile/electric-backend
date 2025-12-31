package facades

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
)

type CotizacionFacade struct {
	cotizacionService *services.CotizacionService
}

func NewCotizacionFacade(cotizacionService *services.CotizacionService) *CotizacionFacade {
	return &CotizacionFacade{
		cotizacionService: cotizacionService,
	}
}

func (f *CotizacionFacade) ObtenerTodas(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.CotizacionModel, int64, error) {
	return f.cotizacionService.ObtenerTodas(ctx, page, limit, filters)
}

func (f *CotizacionFacade) ObtenerPorID(ctx context.Context, id string) (*models.CotizacionModel, error) {
	return f.cotizacionService.ObtenerPorID(ctx, id)
}

func (f *CotizacionFacade) ObtenerPorNumero(ctx context.Context, numero string) (*models.CotizacionModel, error) {
	return f.cotizacionService.ObtenerPorNumero(ctx, numero)
}

func (f *CotizacionFacade) Crear(ctx context.Context, nombre, email, empresa, telefono, servicio, plazo, mensaje string) (*models.CotizacionModel, error) {
	return f.cotizacionService.Crear(ctx, nombre, email, empresa, telefono, servicio, plazo, mensaje)
}

func (f *CotizacionFacade) Actualizar(ctx context.Context, id string, updates map[string]interface{}) (*models.CotizacionModel, error) {
	return f.cotizacionService.Actualizar(ctx, id, updates)
}

func (f *CotizacionFacade) ActualizarEstado(ctx context.Context, id string, estado string) error {
	return f.cotizacionService.ActualizarEstado(ctx, id, estado)
}

func (f *CotizacionFacade) Eliminar(ctx context.Context, id string) error {
	return f.cotizacionService.Eliminar(ctx, id)
}
