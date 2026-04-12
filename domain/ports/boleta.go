package ports

import (
	"context"
	"electric-backend/infrastructure/entities"
)

type PortBoleta interface {
	FindByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error)
	FindByID(ctx context.Context, id string) (*entities.BoletaEntity, error)
	Create(ctx context.Context, boleta *entities.BoletaEntity) error
	Update(ctx context.Context, id string, boleta *entities.BoletaEntity) error
	FindVencidasByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error)
	FindPendientesByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error)
	FindPorVencer(ctx context.Context, diasAntes int) ([]*entities.BoletaEntity, error)
	FindVencidas(ctx context.Context) ([]*entities.BoletaEntity, error)
	FindClienteIDsConBoletasVencidas(ctx context.Context) ([]string, error)
	UpdateEstado(ctx context.Context, id string, estado string) error
	UpdateNotificacionEnviada(ctx context.Context, id string, campo string) error
}

type BoletaRepository interface {
	ObtenerPorCliente(ctx context.Context, clienteID interface{}) ([]*entities.BoletaEntity, error)
}
