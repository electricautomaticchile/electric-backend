package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertaAutomaticaService struct {
	alertaRepo       ports.PortAlerta
	dispositivoRepo  ports.PortDispositivo
	notificacionRepo ports.PortNotificacion
	empresaRepo      ports.PortEmpresa
}

func NewAlertaAutomaticaService(
	alertaRepo ports.PortAlerta,
	dispositivoRepo ports.PortDispositivo,
	notificacionRepo ports.PortNotificacion,
	empresaRepo ports.PortEmpresa,
) *AlertaAutomaticaService {
	return &AlertaAutomaticaService{
		alertaRepo:       alertaRepo,
		dispositivoRepo:  dispositivoRepo,
		notificacionRepo: notificacionRepo,
		empresaRepo:      empresaRepo,
	}
}

func (s *AlertaAutomaticaService) VerificarConsumoAnormal(ctx context.Context, empresaID string) error {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return err
	}

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		consumoActual := dispositivo.UltimaLectura.Energy
		
		if consumoActual > 100 {
			titulo := fmt.Sprintf("Consumo elevado detectado - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s ha registrado un consumo de %.2f kWh, superando el umbral normal.", 
				dispositivo.Nombre, consumoActual)
			
			alerta := &models.AlertaModel{
				Tipo:          "advertencia",
				Titulo:        titulo,
				Mensaje:       mensaje,
				Dispositivo:   dispositivo.NumeroDispositivo,
				EmpresaID:     empresaID,
				Importante:    true,
				Leida:         false,
				Resuelta:      false,
				FechaCreacion: time.Now(),
				Metadatos: map[string]interface{}{
					"consumo":           consumoActual,
					"umbral":            100.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			if err := s.alertaRepo.Create(ctx, alerta); err != nil {
				continue
			}

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				Tipo:           "alerta",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Leida:          false,
				FechaCreacion:  time.Now(),
			}

			s.notificacionRepo.Create(ctx, notificacion)
		}

		if consumoActual < 0.01 && dispositivo.Estado == "activo" {
			titulo := fmt.Sprintf("Posible falla - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s no está registrando consumo. Puede estar desconectado o con falla.", 
				dispositivo.Nombre)
			
			alerta := &models.AlertaModel{
				Tipo:          "advertencia",
				Titulo:        titulo,
				Mensaje:       mensaje,
				Dispositivo:   dispositivo.NumeroDispositivo,
				EmpresaID:     empresaID,
				Importante:    true,
				Leida:         false,
				Resuelta:      false,
				FechaCreacion: time.Now(),
				Metadatos: map[string]interface{}{
					"consumo":           consumoActual,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			if err := s.alertaRepo.Create(ctx, alerta); err != nil {
				continue
			}

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				Tipo:           "alerta",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Leida:          false,
				FechaCreacion:  time.Now(),
			}

			s.notificacionRepo.Create(ctx, notificacion)
		}
	}

	return nil
}

func (s *AlertaAutomaticaService) VerificarPatronesAnormales(ctx context.Context, empresaID string) error {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return err
	}

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		if dispositivo.UltimaLectura.Voltage < 200 || dispositivo.UltimaLectura.Voltage > 240 {
			titulo := fmt.Sprintf("Voltaje anormal - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s registra un voltaje de %.2fV fuera del rango normal (200-240V).", 
				dispositivo.Nombre, dispositivo.UltimaLectura.Voltage)
			
			alerta := &models.AlertaModel{
				Tipo:          "error",
				Titulo:        titulo,
				Mensaje:       mensaje,
				Dispositivo:   dispositivo.NumeroDispositivo,
				EmpresaID:     empresaID,
				Importante:    true,
				Leida:         false,
				Resuelta:      false,
				FechaCreacion: time.Now(),
				Metadatos: map[string]interface{}{
					"voltaje":           dispositivo.UltimaLectura.Voltage,
					"rangoMin":          200.0,
					"rangoMax":          240.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			if err := s.alertaRepo.Create(ctx, alerta); err != nil {
				continue
			}

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				Tipo:           "alerta",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Leida:          false,
				FechaCreacion:  time.Now(),
			}

			s.notificacionRepo.Create(ctx, notificacion)
		}

		if dispositivo.UltimaLectura.Current > 50 {
			titulo := fmt.Sprintf("Corriente elevada - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s registra una corriente de %.2fA, superando el límite seguro.", 
				dispositivo.Nombre, dispositivo.UltimaLectura.Current)
			
			alerta := &models.AlertaModel{
				Tipo:          "error",
				Titulo:        titulo,
				Mensaje:       mensaje,
				Dispositivo:   dispositivo.NumeroDispositivo,
				EmpresaID:     empresaID,
				Importante:    true,
				Leida:         false,
				Resuelta:      false,
				FechaCreacion: time.Now(),
				Metadatos: map[string]interface{}{
					"corriente":         dispositivo.UltimaLectura.Current,
					"umbral":            50.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			if err := s.alertaRepo.Create(ctx, alerta); err != nil {
				continue
			}

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				Tipo:           "alerta",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Leida:          false,
				FechaCreacion:  time.Now(),
			}

			s.notificacionRepo.Create(ctx, notificacion)
		}
	}

	return nil
}

func (s *AlertaAutomaticaService) IniciarMonitoreoAutomatico(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.ejecutarVerificaciones(ctx)
		}
	}
}

func (s *AlertaAutomaticaService) ejecutarVerificaciones(ctx context.Context) {
	empresas, err := s.empresaRepo.FindAll(ctx)
	if err != nil {
		return
	}
	
	for _, empresa := range empresas {
		s.VerificarConsumoAnormal(ctx, empresa.ID)
		s.VerificarPatronesAnormales(ctx, empresa.ID)
	}
}

func (s *AlertaAutomaticaService) VerificarManual(ctx context.Context, empresaID string) error {
	if err := s.VerificarConsumoAnormal(ctx, empresaID); err != nil {
		return err
	}
	return s.VerificarPatronesAnormales(ctx, empresaID)
}
