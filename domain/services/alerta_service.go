package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
)

type AlertaService struct {
	alertaRepo       ports.PortAlerta
	wsNotifier       *WebSocketNotifierService
}

func NewAlertaService(alertaRepo ports.PortAlerta, wsNotifier *WebSocketNotifierService) *AlertaService {
	return &AlertaService{
		alertaRepo:   alertaRepo,
		wsNotifier:   wsNotifier,
	}
}

func (s *AlertaService) ObtenerTodas(ctx context.Context) ([]*models.AlertaModel, error) {
	return s.alertaRepo.FindAll(ctx)
}

func (s *AlertaService) ObtenerActivas(ctx context.Context) ([]*models.AlertaModel, error) {
	return s.alertaRepo.FindActivas(ctx)
}

func (s *AlertaService) ObtenerPorEmpresa(ctx context.Context, empresaID string) ([]*models.AlertaModel, error) {
	return s.alertaRepo.FindByEmpresa(ctx, empresaID)
}

func (s *AlertaService) Crear(ctx context.Context, r *recipe.CrearAlertaRecipe) (*models.AlertaModel, error) {
	model := &models.AlertaModel{
		EmpresaID:   r.EmpresaID,
		Tipo:        r.Tipo,
		Titulo:      r.Mensaje,
		Mensaje:     r.Mensaje,
		Dispositivo: r.DispositivoID,
	}

	if err := s.alertaRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *AlertaService) Resolver(ctx context.Context, id string, r *recipe.ResolverAlertaRecipe) error {
	return s.alertaRepo.Resolver(ctx, id, r.Resolucion)
}

func (s *AlertaService) Eliminar(ctx context.Context, id string) error {
	return s.alertaRepo.Delete(ctx, id)
}
