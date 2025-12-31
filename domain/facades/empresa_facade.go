package facades

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
)

type EmpresaFacade struct {
	empresaService *services.EmpresaService
}

func NewEmpresaFacade(empresaService *services.EmpresaService) *EmpresaFacade {
	return &EmpresaFacade{
		empresaService: empresaService,
	}
}

func (f *EmpresaFacade) ObtenerTodas(ctx context.Context) ([]*models.EmpresaModel, error) {
	return f.empresaService.ObtenerTodas(ctx)
}

func (f *EmpresaFacade) ObtenerPorID(ctx context.Context, id string) (*models.EmpresaModel, error) {
	return f.empresaService.ObtenerPorID(ctx, id)
}

func (f *EmpresaFacade) Crear(ctx context.Context, r *recipe.CrearEmpresaRecipe) (*models.EmpresaModel, error) {
	return f.empresaService.Crear(ctx, r)
}

func (f *EmpresaFacade) Actualizar(ctx context.Context, id string, r *recipe.ActualizarEmpresaRecipe) (*models.EmpresaModel, error) {
	return f.empresaService.Actualizar(ctx, id, r)
}

func (f *EmpresaFacade) Eliminar(ctx context.Context, id string) error {
	return f.empresaService.Eliminar(ctx, id)
}

func (f *EmpresaFacade) CambiarEstado(ctx context.Context, id string, r *recipe.CambiarEstadoEmpresaRecipe) error {
	return f.empresaService.CambiarEstado(ctx, id, r)
}
