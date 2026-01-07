package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/validation"
)

type EmpresaService struct {
	empresaRepo ports.PortEmpresa
}

func NewEmpresaService(empresaRepo ports.PortEmpresa) *EmpresaService {
	return &EmpresaService{
		empresaRepo: empresaRepo,
	}
}

func (s *EmpresaService) ObtenerTodas(ctx context.Context) ([]*models.EmpresaModel, error) {
	return s.empresaRepo.FindAll(ctx)
}

func (s *EmpresaService) ObtenerPorID(ctx context.Context, id string) (*models.EmpresaModel, error) {
	return s.empresaRepo.FindByID(ctx, id)
}

func (s *EmpresaService) Crear(ctx context.Context, r *recipe.CrearEmpresaRecipe) (*models.EmpresaModel, error) {
	rutNormalizado := ""
	if r.RUT != "" {
		rutNormalizado = validation.NormalizarRUT(r.RUT)
	}

	telefonoNormalizado := ""
	if r.Telefono != "" {
		telefonoNormalizado = validation.NormalizarTelefono(r.Telefono)
	}

	model := &models.EmpresaModel{
		NombreEmpresa: validation.SanitizeString(r.Nombre),
		Correo:        validation.SanitizeEmail(r.Email),
		Telefono:      telefonoNormalizado,
		Direccion:     validation.SanitizeString(r.Direccion),
		Rut:           rutNormalizado,
	}

	if err := s.empresaRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *EmpresaService) Actualizar(ctx context.Context, id string, r *recipe.ActualizarEmpresaRecipe) (*models.EmpresaModel, error) {
	empresa, err := s.empresaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.Nombre != "" {
		empresa.NombreEmpresa = validation.SanitizeString(r.Nombre)
	}
	if r.Email != "" {
		empresa.Correo = validation.SanitizeEmail(r.Email)
	}
	if r.Telefono != "" {
		empresa.Telefono = validation.NormalizarTelefono(r.Telefono)
	}
	if r.Direccion != "" {
		empresa.Direccion = validation.SanitizeString(r.Direccion)
	}
	if r.RUT != "" {
		empresa.Rut = validation.NormalizarRUT(r.RUT)
	}

	if err := s.empresaRepo.Update(ctx, id, empresa); err != nil {
		return nil, err
	}

	return empresa, nil
}

func (s *EmpresaService) Eliminar(ctx context.Context, id string) error {
	return s.empresaRepo.Delete(ctx, id)
}

func (s *EmpresaService) CambiarEstado(ctx context.Context, id string, r *recipe.CambiarEstadoEmpresaRecipe) error {
	return s.empresaRepo.CambiarEstado(ctx, id, r.Activo)
}
