package facades

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
)

type AuthFacade struct {
	authService *services.AuthService
}

func NewAuthFacade(authService *services.AuthService) *AuthFacade {
	return &AuthFacade{
		authService: authService,
	}
}

func (f *AuthFacade) Login(ctx context.Context, r *recipe.LoginRecipe) (*models.LoginResponseModel, error) {
	return f.authService.Login(ctx, r)
}

func (f *AuthFacade) ObtenerPerfil(ctx context.Context, userID string) (*models.ClienteModel, error) {
	return f.authService.ObtenerPerfil(ctx, userID)
}

func (f *AuthFacade) CambiarPassword(ctx context.Context, userID string, r *recipe.CambiarPasswordRecipe) error {
	return f.authService.CambiarPassword(ctx, userID, r)
}

func (f *AuthFacade) SolicitarRecuperacion(ctx context.Context, r *recipe.SolicitarRecuperacionRecipe) error {
	return f.authService.SolicitarRecuperacion(ctx, r)
}

func (f *AuthFacade) RestablecerPassword(ctx context.Context, r *recipe.RestablecerPasswordRecipe) error {
	return f.authService.RestablecerPassword(ctx, r)
}

func (f *AuthFacade) RegistrarEmpresa(ctx context.Context, r *recipe.RegistroEmpresaRecipe) (*models.EmpresaModel, error) {
	return f.authService.RegistrarEmpresa(ctx, r)
}
