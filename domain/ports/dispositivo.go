package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortDispositivo interface {
	FindAll(ctx context.Context, empresaID string) ([]*entities.DispositivoEntity, error)
	FindByID(ctx context.Context, id string) (*entities.DispositivoEntity, error)
	FindByNumero(ctx context.Context, numeroDispositivo string) (*entities.DispositivoEntity, error)
	FindByCliente(ctx context.Context, clienteID string) ([]*entities.DispositivoEntity, error)
	Create(ctx context.Context, dispositivo *entities.DispositivoEntity) error
	Update(ctx context.Context, id string, dispositivo *entities.DispositivoEntity) error
	UpdateUltimaLectura(ctx context.Context, numeroDispositivo string, lectura *entities.LecturaDispositivo) error
	CambiarEstado(ctx context.Context, id string, estado string) error
	Delete(ctx context.Context, id string) error
}
