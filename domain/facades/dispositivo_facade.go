package facades

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
)

type DispositivoFacade struct {
	dispositivoService *services.DispositivoService
}

func NewDispositivoFacade(dispositivoService *services.DispositivoService) *DispositivoFacade {
	return &DispositivoFacade{
		dispositivoService: dispositivoService,
	}
}

func (f *DispositivoFacade) ObtenerTodos(ctx context.Context, empresaID string) ([]*models.DispositivoModel, error) {
	return f.dispositivoService.ObtenerTodos(ctx, empresaID)
}

func (f *DispositivoFacade) ObtenerPorID(ctx context.Context, id string) (*models.DispositivoModel, error) {
	return f.dispositivoService.ObtenerPorID(ctx, id)
}

func (f *DispositivoFacade) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*models.DispositivoModel, error) {
	return f.dispositivoService.ObtenerPorCliente(ctx, clienteID)
}

func (f *DispositivoFacade) Crear(ctx context.Context, r *recipe.CrearDispositivoRecipe) (*models.DispositivoModel, error) {
	return f.dispositivoService.Crear(ctx, r)
}

func (f *DispositivoFacade) Actualizar(ctx context.Context, id string, r *recipe.ActualizarDispositivoRecipe) (*models.DispositivoModel, error) {
	return f.dispositivoService.Actualizar(ctx, id, r)
}

func (f *DispositivoFacade) AsignarCliente(ctx context.Context, id string, clienteID string) (*models.DispositivoModel, error) {
	return f.dispositivoService.AsignarCliente(ctx, id, clienteID)
}

func (f *DispositivoFacade) ActualizarUltimaLectura(ctx context.Context, numeroDispositivo string, r *recipe.ActualizarLecturaRecipe) error {
	return f.dispositivoService.ActualizarUltimaLectura(ctx, numeroDispositivo, r)
}

func (f *DispositivoFacade) CambiarEstado(ctx context.Context, id string, r *recipe.CambiarEstadoDispositivoRecipe) error {
	return f.dispositivoService.CambiarEstado(ctx, id, r)
}

func (f *DispositivoFacade) Eliminar(ctx context.Context, id string) error {
	return f.dispositivoService.Eliminar(ctx, id)
}
