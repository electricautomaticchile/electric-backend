package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FCMTokenService struct {
	fcmTokenRepo ports.PortFCMToken
}

func NewFCMTokenService(fcmTokenRepo ports.PortFCMToken) *FCMTokenService {
	return &FCMTokenService{
		fcmTokenRepo: fcmTokenRepo,
	}
}

// RegistrarToken registra o actualiza (upsert idempotente) un token FCM para el
// usuario indicado. Permite múltiples tokens por usuario (multi-dispositivo).
func (s *FCMTokenService) RegistrarToken(ctx context.Context, userID, token, plataforma string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return types.ThrowData("ID de usuario inválido")
	}

	token = validation.SanitizeString(token)
	if token == "" {
		return types.ThrowRecipe("El token es requerido", "token")
	}

	entity := &entities.FCMTokenEntity{
		UserID:     userObjectID,
		Token:      token,
		Plataforma: plataforma,
	}

	return s.fcmTokenRepo.Upsert(ctx, entity)
}

// ObtenerTokensDeUsuario devuelve la lista de tokens FCM de un usuario.
func (s *FCMTokenService) ObtenerTokensDeUsuario(ctx context.Context, userID string) ([]*models.FCMTokenModel, error) {
	tokens, err := s.fcmTokenRepo.FindByUsuario(ctx, userID)
	if err != nil {
		return []*models.FCMTokenModel{}, nil
	}

	result := make([]*models.FCMTokenModel, len(tokens))
	for i, token := range tokens {
		result[i] = s.entityToModel(token)
	}

	return result, nil
}

// EliminarToken elimina un token FCM por su valor.
func (s *FCMTokenService) EliminarToken(ctx context.Context, token string) error {
	return s.fcmTokenRepo.DeleteByToken(ctx, token)
}

func (s *FCMTokenService) entityToModel(entity *entities.FCMTokenEntity) *models.FCMTokenModel {
	return &models.FCMTokenModel{
		ID:         entity.ID.Hex(),
		UserID:     entity.UserID.Hex(),
		Token:      entity.Token,
		Plataforma: entity.Plataforma,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}
}
