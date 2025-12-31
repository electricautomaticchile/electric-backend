package facades

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
)

type ClienteFacade struct {
	clienteService *services.ClienteService
}

func NewClienteFacade(clienteService *services.ClienteService) *ClienteFacade {
	return &ClienteFacade{
		clienteService: clienteService,
	}
}

func (f *ClienteFacade) ObtenerTodos(ctx context.Context, empresaID string) ([]*models.ClienteModel, error) {
	return f.clienteService.ObtenerTodos(ctx, empresaID)
}

func (f *ClienteFacade) ObtenerPorID(ctx context.Context, id string) (*models.ClienteModel, error) {
	return f.clienteService.ObtenerPorID(ctx, id)
}

func (f *ClienteFacade) ObtenerPorNumero(ctx context.Context, numeroCliente string) (*models.ClienteModel, error) {
	return f.clienteService.ObtenerPorNumero(ctx, numeroCliente)
}

func (f *ClienteFacade) Crear(ctx context.Context, r *recipe.CrearClienteRecipe) (*models.ClienteModel, error) {
	return f.clienteService.Crear(ctx, r)
}

func (f *ClienteFacade) Actualizar(ctx context.Context, id string, r *recipe.ActualizarClienteRecipe) (*models.ClienteModel, error) {
	return f.clienteService.Actualizar(ctx, id, r)
}

func (f *ClienteFacade) Eliminar(ctx context.Context, id string) error {
	return f.clienteService.Eliminar(ctx, id)
}
