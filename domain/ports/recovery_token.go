package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortRecoveryToken interface {
	Create(ctx context.Context, token *entities.RecoveryTokenEntity) error
	FindByToken(ctx context.Context, token string) (*entities.RecoveryTokenEntity, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
	DeleteExpired(ctx context.Context) error
}
