package services

import (
"context"
"electric-backend/domain/ports"
)

type EstadisticaService struct {
	estadisticaRepo ports.PortEstadistica
}

func NewEstadisticaService(estadisticaRepo ports.PortEstadistica) *EstadisticaService {
	return &EstadisticaService{
		estadisticaRepo: estadisticaRepo,
	}
}

func (s *EstadisticaService) ObtenerConsumoCliente(ctx context.Context, clienteID string) (map[string]interface{}, error) {
	return s.estadisticaRepo.ObtenerConsumoCliente(ctx, clienteID)
}

func (s *EstadisticaService) ObtenerEstadisticasGlobales(ctx context.Context) (map[string]interface{}, error) {
	return s.estadisticaRepo.ObtenerEstadisticasGlobales(ctx)
}
