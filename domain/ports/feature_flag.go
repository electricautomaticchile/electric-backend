package ports

import (
	"context"

	"electric-backend/infrastructure/entities"
)

type PortFeatureFlag interface {
	// FindAll devuelve todos los flags.
	FindAll(ctx context.Context) ([]*entities.FeatureFlagEntity, error)
	// FindByKey devuelve un flag por su clave (nil si no existe).
	FindByKey(ctx context.Context, key string) (*entities.FeatureFlagEntity, error)
	// Upsert crea o actualiza un flag de forma idempotente por key.
	Upsert(ctx context.Context, flag *entities.FeatureFlagEntity) error
	// Delete elimina un flag por su key.
	Delete(ctx context.Context, key string) error
}
