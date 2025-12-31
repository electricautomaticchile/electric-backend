package services

import (
"context"
"electric-backend/api/v1/recipe"
"electric-backend/domain/models"
"electric-backend/domain/ports"
"electric-backend/infrastructure/entities"
)

type ConfiguracionService struct {
	configuracionRepo ports.PortConfiguracion
}

func NewConfiguracionService(configuracionRepo ports.PortConfiguracion) *ConfiguracionService {
	return &ConfiguracionService{
		configuracionRepo: configuracionRepo,
	}
}

func (s *ConfiguracionService) ObtenerTodas(ctx context.Context) ([]*models.ConfiguracionModel, error) {
	configuraciones, err := s.configuracionRepo.FindAll(ctx)
	if err != nil {
		return []*models.ConfiguracionModel{}, nil
	}

	models := make([]*models.ConfiguracionModel, len(configuraciones))
	for i, config := range configuraciones {
		models[i] = s.entityToModel(config)
	}

	return models, nil
}

func (s *ConfiguracionService) ObtenerPorClave(ctx context.Context, clave string) (*models.ConfiguracionModel, error) {
	config, err := s.configuracionRepo.FindByClave(ctx, clave)
	if err != nil {
		return nil, err
	}

	return s.entityToModel(config), nil
}

func (s *ConfiguracionService) Actualizar(ctx context.Context, r *recipe.ActualizarConfiguracionRecipe) error {
	return s.configuracionRepo.Actualizar(ctx, r.Clave, r.Valor)
}

func (s *ConfiguracionService) entityToModel(entity *entities.ConfiguracionEntity) *models.ConfiguracionModel {
	return &models.ConfiguracionModel{
		ID:                 entity.ID.Hex(),
		Clave:              entity.Clave,
		Valor:              entity.Valor,
		Categoria:          entity.Categoria,
		FechaActualizacion: entity.FechaActualizacion,
	}
}
