package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/sms"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificacionSMSService struct {
	clienteRepo     ports.PortCliente
	boletaRepo      ports.BoletaRepository
	dispositivoRepo ports.DispositivoRepository
	smsService      sms.SMSService
}

func NewNotificacionSMSService(
	clienteRepo ports.PortCliente,
	boletaRepo ports.BoletaRepository,
	dispositivoRepo ports.DispositivoRepository,
	smsService sms.SMSService,
) *NotificacionSMSService {
	return &NotificacionSMSService{
		clienteRepo:     clienteRepo,
		boletaRepo:      boletaRepo,
		dispositivoRepo: dispositivoRepo,
		smsService:      smsService,
	}
}

func (s *NotificacionSMSService) EnviarNotificacionesConsumoQuincenal(ctx context.Context) error {
	clientes, err := s.clienteRepo.FindAll(ctx, "")
	if err != nil {
		return fmt.Errorf("error obteniendo clientes: %w", err)
	}

	ahora := time.Now()
	for _, cliente := range clientes {
		if !cliente.NotificacionesSMS || cliente.Telefono == "" {
			continue
		}

		dispositivos, err := s.dispositivoRepo.ObtenerPorCliente(ctx, cliente.ID)
		if err != nil || len(dispositivos) == 0 {
			continue
		}

		var consumoTotal float64
		var costoTotal float64

		for _, dispositivo := range dispositivos {
			if dispositivo.UltimaLectura != nil {
				consumoTotal += dispositivo.UltimaLectura.ConsumoKWh
				costoTotal += dispositivo.UltimaLectura.CostoEstimado
			}
		}

		if consumoTotal > 0 {
			diasTranscurridos := 15
			if cliente.FechaActivacion != nil {
				diasDesdeActivacion := int(ahora.Sub(*cliente.FechaActivacion).Hours() / 24)
				if diasDesdeActivacion < 15 {
					diasTranscurridos = diasDesdeActivacion
				}
			}

			err = s.smsService.EnviarNotificacionConsumo(
				cliente.Telefono,
				cliente.Nombre,
				consumoTotal,
				costoTotal,
				diasTranscurridos,
			)
			if err != nil {
				fmt.Printf("Error enviando SMS a %s: %v\n", cliente.Nombre, err)
			}
		}
	}

	return nil
}

func (s *NotificacionSMSService) VerificarYNotificarBoletasImpagas(ctx context.Context) error {
	clientes, err := s.clienteRepo.FindAll(ctx, "")
	if err != nil {
		return fmt.Errorf("error obteniendo clientes: %w", err)
	}

	for _, cliente := range clientes {
		if !cliente.NotificacionesSMS || cliente.Telefono == "" {
			continue
		}

		clienteID, err := primitive.ObjectIDFromHex(cliente.ID)
		if err != nil {
			continue
		}

		boletas, err := s.boletaRepo.ObtenerPorCliente(ctx, clienteID)
		if err != nil {
			continue
		}

		var boletasImpagas []interface{}
		var montoTotal float64

		for _, boleta := range boletas {
			if boleta.Estado == "pendiente" || boleta.Estado == "por_vencer" || boleta.Estado == "vencido" {
				boletasImpagas = append(boletasImpagas, boleta)
				montoTotal += boleta.MontoTotal
			}
		}

		if len(boletasImpagas) >= 3 {
			err = s.smsService.EnviarAvisoCorteServicio(
				cliente.Telefono,
				cliente.Nombre,
				len(boletasImpagas),
				montoTotal,
			)
			if err != nil {
				fmt.Printf("Error enviando SMS de corte a %s: %v\n", cliente.Nombre, err)
			}
		}
	}

	return nil
}

func (s *NotificacionSMSService) NotificarPagoBoleta(ctx context.Context, clienteID, periodo string, monto float64) error {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return fmt.Errorf("error obteniendo cliente: %w", err)
	}

	if !cliente.NotificacionesSMS || cliente.Telefono == "" {
		return nil
	}

	return s.smsService.EnviarConfirmacionPago(
		cliente.Telefono,
		cliente.Nombre,
		periodo,
		monto,
	)
}
