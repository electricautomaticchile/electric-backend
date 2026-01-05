package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/entities"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoletaService struct {
	boletaRepo   ports.PortBoleta
	clienteRepo  ports.PortCliente
	emailService *email.ResendService
}

func NewBoletaService(boletaRepo ports.PortBoleta, clienteRepo ports.PortCliente, emailService *email.ResendService) *BoletaService {
	return &BoletaService{
		boletaRepo:   boletaRepo,
		clienteRepo:  clienteRepo,
		emailService: emailService,
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

	go s.enviarBoletaPorEmail(ctx, r.ClienteID, entity)

	return s.entityToModel(entity), nil
}

func (s *BoletaService) enviarBoletaPorEmail(ctx context.Context, clienteID string, boleta *entities.BoletaEntity) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		log.Printf("Error obteniendo cliente para email de boleta: %v", err)
		return
	}

	if cliente.Correo == "" {
		return
	}

	numeroBoleta := boleta.ID.Hex()[:8]
	monto := fmt.Sprintf("%.2f", boleta.Monto)
	fechaVencimiento := time.Now().AddDate(0, 0, 30).Format("02/01/2006")

	err = s.emailService.EnviarNotificacionBoleta(
		cliente.Correo,
		cliente.Nombre,
		numeroBoleta,
		monto,
		fechaVencimiento,
	)

	if err != nil {
		log.Printf("Error enviando email de boleta: %v", err)
	}
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
