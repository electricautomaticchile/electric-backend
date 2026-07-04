package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortFCMToken interface {
	// Upsert registra o actualiza un token de forma idempotente por token.
	Upsert(ctx context.Context, token *entities.FCMTokenEntity) error
	// FindByUsuario devuelve todos los tokens registrados para un usuario.
	FindByUsuario(ctx context.Context, userID string) ([]*entities.FCMTokenEntity, error)
	// DeleteByToken elimina un token por su valor.
	DeleteByToken(ctx context.Context, token string) error
}
