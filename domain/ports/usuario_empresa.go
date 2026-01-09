package ports

import (
	"context"
	"electric-backend/domain/models"
)

type PortUsuarioEmpresa interface {
	Create(ctx context.Context, usuario *models.UsuarioEmpresaModel) error
	FindAll(ctx context.Context, empresaID string) ([]*models.UsuarioEmpresaModel, error)
	FindByID(ctx context.Context, id string) (*models.UsuarioEmpresaModel, error)
	FindByEmail(ctx context.Context, email string) (*models.UsuarioEmpresaModel, error)
	Update(ctx context.Context, id string, usuario *models.UsuarioEmpresaModel) error
	Delete(ctx context.Context, id string) error
	UpdateUltimoAcceso(ctx context.Context, id string) error
}
