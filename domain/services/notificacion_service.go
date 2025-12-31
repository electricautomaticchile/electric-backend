package services

import (
"context"
"electric-backend/api/v1/recipe"
"electric-backend/domain/models"
"electric-backend/domain/ports"
"electric-backend/infrastructure/entities"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificacionService struct {
	notificacionRepo ports.PortNotificacion
}

func NewNotificacionService(notificacionRepo ports.PortNotificacion) *NotificacionService {
	return &NotificacionService{
		notificacionRepo: notificacionRepo,
	}
}

func (s *NotificacionService) Listar(ctx context.Context, destinatarioID string) ([]*models.NotificacionModel, error) {
	notificaciones, err := s.notificacionRepo.FindByDestinatario(ctx, destinatarioID)
	if err != nil {
		return []*models.NotificacionModel{}, nil
	}

	models := make([]*models.NotificacionModel, len(notificaciones))
	for i, notificacion := range notificaciones {
		models[i] = s.entityToModel(notificacion)
	}

	return models, nil
}

func (s *NotificacionService) MarcarLeida(ctx context.Context, id string) error {
	return s.notificacionRepo.MarcarLeida(ctx, id)
}

func (s *NotificacionService) MarcarTodasLeidas(ctx context.Context, destinatarioID string) error {
	return s.notificacionRepo.MarcarTodasLeidas(ctx, destinatarioID)
}

func (s *NotificacionService) Eliminar(ctx context.Context, id string) error {
	return s.notificacionRepo.Delete(ctx, id)
}

func (s *NotificacionService) Crear(ctx context.Context, r *recipe.CrearNotificacionRecipe) (*models.NotificacionModel, error) {
	destinatarioID, _ := primitive.ObjectIDFromHex(r.DestinatarioID)

	entity := &entities.NotificacionEntity{
		DestinatarioID: destinatarioID,
		Titulo:         r.Titulo,
		Mensaje:        r.Mensaje,
		Tipo:           r.Tipo,
	}

	if err := s.notificacionRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return s.entityToModel(entity), nil
}

func (s *NotificacionService) entityToModel(entity *entities.NotificacionEntity) *models.NotificacionModel {
	return &models.NotificacionModel{
		ID:             entity.ID.Hex(),
		DestinatarioID: entity.DestinatarioID.Hex(),
		Titulo:         entity.Titulo,
		Mensaje:        entity.Mensaje,
		Tipo:           entity.Tipo,
		Leida:          entity.Leida,
		FechaCreacion:  entity.FechaCreacion,
	}
}
