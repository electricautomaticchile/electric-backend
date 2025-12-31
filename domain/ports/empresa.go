package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortEmpresa interface {
	FindAll(ctx context.Context) ([]*models.EmpresaModel, error)
	FindByID(ctx context.Context, id string) (*models.EmpresaModel, error)
	FindByNumeroCliente(ctx context.Context, numeroCliente string) (*models.EmpresaModel, error)
	Create(ctx context.Context, empresa *models.EmpresaModel) error
	Update(ctx context.Context, id string, empresa *models.EmpresaModel) error
	UpdateUltimoAcceso(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id string, hashedPassword string) error
	Delete(ctx context.Context, id string) error
	CambiarEstado(ctx context.Context, id string, activo bool) error
}
