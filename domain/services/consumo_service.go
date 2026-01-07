package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"time"
)

type ConsumoService struct {
	clienteRepo ports.PortCliente
	tarifaRepo  ports.PortTarifa
}

func NewConsumoService(clienteRepo ports.PortCliente, tarifaRepo ports.PortTarifa) *ConsumoService {
	return &ConsumoService{
		clienteRepo: clienteRepo,
		tarifaRepo:  tarifaRepo,
	}
}

func (s *ConsumoService) CalcularCostoLectura(ctx context.Context, clienteID string, kwhConsumidos float64) (*models.CalculoConsumoResponse, error) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return nil, err
	}

	var tarifa *models.TarifaModel

	if cliente.TarifaID != "" {
		tarifa, err = s.tarifaRepo.FindByID(ctx, cliente.TarifaID)
		if err != nil {
			return nil, err
		}
	} else if cliente.Comuna != "" && cliente.TipoTarifa != "" {
		tarifa, err = s.tarifaRepo.FindActiva(ctx, cliente.Comuna, cliente.TipoTarifa)
		if err != nil {
			return nil, err
		}
	} else {
		tarifa, err = s.tarifaRepo.FindActiva(ctx, "Villa Alemana", "BT1")
		if err != nil {
			return nil, err
		}
	}

	return s.calcularMonto(kwhConsumidos, tarifa), nil
}

func (s *ConsumoService) CalcularConsumoMensual(ctx context.Context, clienteID string, mes time.Time) (*models.CalculoConsumoResponse, error) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return nil, err
	}

	var tarifa *models.TarifaModel
	if cliente.TarifaID != "" {
		tarifa, err = s.tarifaRepo.FindByID(ctx, cliente.TarifaID)
	} else if cliente.Comuna != "" && cliente.TipoTarifa != "" {
		tarifa, err = s.tarifaRepo.FindActiva(ctx, cliente.Comuna, cliente.TipoTarifa)
	} else {
		tarifa, err = s.tarifaRepo.FindActiva(ctx, "Villa Alemana", "BT1")
	}

	if err != nil {
		return nil, err
	}

	return &models.CalculoConsumoResponse{
		PrecioKwhBase: tarifa.PrecioKwhBase,
	}, nil
}

func (s *ConsumoService) calcularMonto(kwhConsumidos float64, tarifa *models.TarifaModel) *models.CalculoConsumoResponse {
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
	}
}
