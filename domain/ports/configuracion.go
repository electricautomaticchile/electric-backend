package ports

import (
"context"
"electric-backend/infrastructure/entities"
)

type PortConfiguracion interface {
	FindAll(ctx context.Context) ([]*entities.ConfiguracionEntity, error)
	FindByClave(ctx context.Context, clave string) (*entities.ConfiguracionEntity, error)
	Actualizar(ctx context.Context, clave string, valor interface{}) error
}
