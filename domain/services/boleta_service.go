package services

import (
"context"
"electric-backend/api/v1/recipe"
"electric-backend/domain/models"
"electric-backend/domain/ports"
"electric-backend/infrastructure/entities"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoletaService struct {
	boletaRepo ports.PortBoleta
}

func NewBoletaService(boletaRepo ports.PortBoleta) *BoletaService {
	return &BoletaService{
		boletaRepo: boletaRepo,
	}
}

func (s *BoletaService) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*models.BoletaModel, error) {
	boletas, err := s.boletaRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return []*models.BoletaModel{}, nil
	}

	models := make([]*models.BoletaModel, len(boletas))
	for i, boleta := range boletas {
		models[i] = s.entityToModel(boleta)
	}

	return models, nil
}

func (s *BoletaService) ObtenerPorID(ctx context.Context, id string) (*models.BoletaModel, error) {
	boleta, err := s.boletaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToModel(boleta), nil
}

func (s *BoletaService) Crear(ctx context.Context, r *recipe.CrearBoletaRecipe) (*models.BoletaModel, error) {
	clienteID, _ := primitive.ObjectIDFromHex(r.ClienteID)

	entity := &entities.BoletaEntity{
		ClienteID: clienteID,
		Monto:     r.Monto,
		Periodo:   r.Periodo,
	}

	if err := s.boletaRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return s.entityToModel(entity), nil
}

func (s *BoletaService) entityToModel(entity *entities.BoletaEntity) *models.BoletaModel {
	return &models.BoletaModel{
		ID:            entity.ID.Hex(),
		ClienteID:     entity.ClienteID.Hex(),
		Monto:         entity.Monto,
		Periodo:       entity.Periodo,
		Estado:        entity.Estado,
		FechaCreacion: entity.FechaCreacion,
		FechaPago:     entity.FechaPago,
	}
}
