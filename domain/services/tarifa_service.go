package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
)

type TarifaService struct {
	tarifaRepo ports.PortTarifa
}

func NewTarifaService(tarifaRepo ports.PortTarifa) *TarifaService {
	return &TarifaService{
		tarifaRepo: tarifaRepo,
	}
}

func (s *TarifaService) CrearTarifa(ctx context.Context, tarifa *models.TarifaModel) error {
	return s.tarifaRepo.Create(ctx, tarifa)
}

func (s *TarifaService) ObtenerTarifa(ctx context.Context, id string) (*models.TarifaModel, error) {
	return s.tarifaRepo.FindByID(ctx, id)
}

func (s *TarifaService) ObtenerTarifaActiva(ctx context.Context, comuna, tipoTarifa string) (*models.TarifaModel, error) {
	return s.tarifaRepo.FindActiva(ctx, comuna, tipoTarifa)
}

func (s *TarifaService) ObtenerTodasTarifas(ctx context.Context) ([]*models.TarifaModel, error) {
	return s.tarifaRepo.FindAll(ctx)
}

func (s *TarifaService) ActualizarTarifa(ctx context.Context, id string, tarifa *models.TarifaModel) error {
	return s.tarifaRepo.Update(ctx, id, tarifa)
}

func (s *TarifaService) EliminarTarifa(ctx context.Context, id string) error {
	return s.tarifaRepo.Delete(ctx, id)
}

func (s *TarifaService) CalcularMontoConsumo(ctx context.Context, kwhConsumidos float64, tarifaID string) (*models.CalculoConsumoResponse, error) {
	tarifa, err := s.tarifaRepo.FindByID(ctx, tarifaID)
	if err != nil {
		return nil, err
	}

	montoBase := kwhConsumidos * tarifa.PrecioKwhBase
	
	var cargoEstabilizacion float64
	var montoEstabilizacion float64
	var tramo string

	if kwhConsumidos > 500 {
		cargoEstabilizacion = tarifa.TramosEstabilizacion.Mayor500Kwh
		montoEstabilizacion = kwhConsumidos * cargoEstabilizacion
		tramo = "Mayor a 500 kWh"
	} else if kwhConsumidos > 350 {
		cargoEstabilizacion = tarifa.TramosEstabilizacion.Entre350Y500
		montoEstabilizacion = kwhConsumidos * cargoEstabilizacion
		tramo = "Entre 350 y 500 kWh"
	} else {
		cargoEstabilizacion = tarifa.TramosEstabilizacion.Hasta350Kwh
		montoEstabilizacion = 0
		tramo = "Hasta 350 kWh"
	}

	montoTotal := montoBase + montoEstabilizacion

	return &models.CalculoConsumoResponse{
		KwhConsumidos:       kwhConsumidos,
		PrecioKwhBase:       tarifa.PrecioKwhBase,
		CargoEstabilizacion: cargoEstabilizacion,
		MontoBase:           montoBase,
		MontoEstabilizacion: montoEstabilizacion,
		MontoTotal:          montoTotal,
		TramoEstabilizacion: tramo,
	}, nil
}
