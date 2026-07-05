package ports

import (
	"context"

	"electric-backend/infrastructure/entities"
)

type PortUmbralAlerta interface {
	// FindByEmpresa devuelve el umbral configurado de una empresa (nil si no existe).
	FindByEmpresa(ctx context.Context, empresaID string) (*entities.UmbralAlertaEntity, error)
	// Upsert crea o actualiza el umbral de una empresa de forma idempotente por empresaId.
	Upsert(ctx context.Context, umbral *entities.UmbralAlertaEntity) error
}
