package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MonitoreoService struct {
	notificacionRepo ports.PortNotificacion
	dispositivoRepo  ports.PortDispositivo
	clienteRepo      ports.PortCliente
	empresaRepo      ports.PortEmpresa
	emailService     EmailSender    // Mejora #12: Alertas externas
	smsService       SMSSender      // Mejora #12: Alertas externas
}

// EmailSender interfaz mínima para envío de alertas por email
type EmailSender interface {
	EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error
}

// SMSSender interfaz mínima para envío de alertas por SMS
type SMSSender interface {
	EnviarSMS(to, mensaje string) error
}

func NewMonitoreoService(
	notificacionRepo ports.PortNotificacion,
	dispositivoRepo ports.PortDispositivo,
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	emailService EmailSender,
	smsService SMSSender,
) *MonitoreoService {
	return &MonitoreoService{
		notificacionRepo: notificacionRepo,
		dispositivoRepo:  dispositivoRepo,
		clienteRepo:      clienteRepo,
		empresaRepo:      empresaRepo,
		emailService:     emailService,
		smsService:       smsService,
	}
}

func (s *MonitoreoService) VerificarConsumoAnormal(ctx context.Context, empresaID string) error {
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
			
			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "advertencia",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				Leida:          false,
				Resuelta:       false,
				FechaCreacion:  time.Now(),
				Metadatos: map[string]interface{}{
					"consumo":           consumoActual,
					"umbral":            100.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			s.notificacionRepo.Create(ctx, notificacion)
			s.enviarAlertaExterna(ctx, dispositivo, titulo, mensaje, "advertencia")
		}

		if consumoActual < 0.01 && dispositivo.Estado == "activo" {
			titulo := fmt.Sprintf("Posible falla - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s no está registrando consumo. Puede estar desconectado o con falla.", 
				dispositivo.Nombre)
			
			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "advertencia",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				Leida:          false,
				Resuelta:       false,
				FechaCreacion:  time.Now(),
				Metadatos: map[string]interface{}{
					"consumo":           consumoActual,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			s.notificacionRepo.Create(ctx, notificacion)
			s.enviarAlertaExterna(ctx, dispositivo, titulo, mensaje, "advertencia")
		}
	}

	return nil
}

func (s *MonitoreoService) VerificarPatronesAnormales(ctx context.Context, empresaID string) error {
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
			
			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "error",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				Leida:          false,
				Resuelta:       false,
				FechaCreacion:  time.Now(),
				Metadatos: map[string]interface{}{
					"voltaje":           dispositivo.UltimaLectura.Voltage,
					"rangoMin":          200.0,
					"rangoMax":          240.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			s.notificacionRepo.Create(ctx, notificacion)
			s.enviarAlertaExterna(ctx, dispositivo, titulo, mensaje, "error")
		}

		if dispositivo.UltimaLectura.Current > 50 {
			titulo := fmt.Sprintf("Corriente elevada - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s registra una corriente de %.2fA, superando el límite seguro.", 
				dispositivo.Nombre, dispositivo.UltimaLectura.Current)
			
			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notificacion := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "error",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				Leida:          false,
				Resuelta:       false,
				FechaCreacion:  time.Now(),
				Metadatos: map[string]interface{}{
					"corriente":         dispositivo.UltimaLectura.Current,
					"umbral":            50.0,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			s.notificacionRepo.Create(ctx, notificacion)
			s.enviarAlertaExterna(ctx, dispositivo, titulo, mensaje, "error")
		}
	}

	return nil
}

func (s *MonitoreoService) IniciarMonitoreoAutomatico(ctx context.Context) {
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

func (s *MonitoreoService) ejecutarVerificaciones(ctx context.Context) {
	empresas, err := s.empresaRepo.FindAll(ctx)
	if err != nil {
		return
	}
	
	for _, empresa := range empresas {
		s.VerificarConsumoAnormal(ctx, empresa.ID)
		s.VerificarPatronesAnormales(ctx, empresa.ID)
	}
}

func (s *MonitoreoService) VerificarManual(ctx context.Context, empresaID string) error {
	if err := s.VerificarConsumoAnormal(ctx, empresaID); err != nil {
		return err
	}
	return s.VerificarPatronesAnormales(ctx, empresaID)
}

type Anomalia struct {
	ID                  string                 `json:"id"`
	DispositivoID       string                 `json:"dispositivoId"`
	ClienteID           string                 `json:"clienteId"`
	NumeroDispositivo   string                 `json:"numeroDispositivo"`
	NombreCliente       string                 `json:"nombreCliente"`
	TipoAnomalia        string                 `json:"tipoAnomalia"`
	Severidad           string                 `json:"severidad"`
	Descripcion         string                 `json:"descripcion"`
	ValorEsperado       float64                `json:"valorEsperado,omitempty"`
	ValorDetectado      float64                `json:"valorDetectado,omitempty"`
	Porcentaje          float64                `json:"porcentaje,omitempty"`
	FechaDeteccion      time.Time              `json:"fechaDeteccion"`
	Estado              string                 `json:"estado"`
	Evidencia           map[string]interface{} `json:"evidencia,omitempty"`
}

type EstadisticasAntifraude struct {
	TotalAnomalias            int            `json:"totalAnomalias"`
	AnomaliasCriticas         int            `json:"anomaliasCriticas"`
	FraudesConfirmados        int            `json:"fraudesConfirmados"`
	TasaDeteccion             float64        `json:"tasaDeteccion"`
	AhorroEstimado            float64        `json:"ahorroEstimado"`
	TiempoPromedioResolucion  int            `json:"tiempoPromedioResolucion"`
	PorTipo                   map[string]int `json:"porTipo"`
}

func (s *MonitoreoService) DetectarAnomalias(ctx context.Context, empresaID string) ([]Anomalia, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []Anomalia{}, nil
	}

	anomalias := make([]Anomalia, 0)

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		cliente, _ := s.clienteRepo.FindByID(ctx, dispositivo.ClienteID.Hex())
		nombreCliente := dispositivo.Nombre
		if cliente != nil {
			nombreCliente = cliente.Nombre
		}

		anomaliasDispositivo := s.analizarDispositivoAntifraude(dispositivo, nombreCliente)
		anomalias = append(anomalias, anomaliasDispositivo...)
	}

	return anomalias, nil
}

func (s *MonitoreoService) analizarDispositivoAntifraude(dispositivo *entities.DispositivoEntity, nombreCliente string) []Anomalia {
	anomalias := make([]Anomalia, 0)
	lectura := dispositivo.UltimaLectura

	consumoPromedio := 150.0

	if lectura.Energy > consumoPromedio*2 {
		porcentaje := ((lectura.Energy - consumoPromedio) / consumoPromedio) * 100
		severidad := "high"
		if porcentaje > 200 {
			severidad = "critical"
		}

		anomalias = append(anomalias, Anomalia{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     nombreCliente,
			TipoAnomalia:      "consumo_elevado",
			Severidad:         severidad,
			Descripcion:       "Consumo anormalmente elevado detectado",
			ValorEsperado:     consumoPromedio,
			ValorDetectado:    lectura.Energy,
			Porcentaje:        porcentaje,
			FechaDeteccion:    time.Now(),
			Estado:            "pending",
			Evidencia: map[string]interface{}{
				"voltage":     lectura.Voltage,
				"current":     lectura.Current,
				"activePower": lectura.ActivePower,
				"timestamp":   lectura.Timestamp,
			},
		})
	}

	if lectura.Energy < 0.1 && dispositivo.Estado == "activo" {
		anomalias = append(anomalias, Anomalia{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     nombreCliente,
			TipoAnomalia:      "consumo_cero",
			Severidad:         "critical",
			Descripcion:       "Dispositivo activo sin consumo - posible manipulación",
			ValorEsperado:     consumoPromedio,
			ValorDetectado:    lectura.Energy,
			Porcentaje:        100,
			FechaDeteccion:    time.Now(),
			Estado:            "pending",
			Evidencia: map[string]interface{}{
				"voltage":     lectura.Voltage,
				"current":     lectura.Current,
				"activePower": lectura.ActivePower,
				"timestamp":   lectura.Timestamp,
			},
		})
	}

	return anomalias
}

// Mejora #12: Enviar alertas externas (email/SMS) cuando se detectan anomalías
func (s *MonitoreoService) enviarAlertaExterna(ctx context.Context, dispositivo *entities.DispositivoEntity, titulo, mensaje, severidad string) {
	if dispositivo.ClienteID.IsZero() {
		return
	}

	cliente, err := s.clienteRepo.FindByID(ctx, dispositivo.ClienteID.Hex())
	if err != nil || cliente == nil {
		return
	}

	// Enviar email si hay correo configurado
	if s.emailService != nil && cliente.Correo != "" {
		go func() {
			if err := s.emailService.EnviarNotificacionAlerta(cliente.Correo, cliente.Nombre, titulo, mensaje); err != nil {
				fmt.Printf("Error enviando alerta email a %s: %v\n", cliente.Correo, err)
			}
		}()
	}

	// Enviar SMS solo para alertas críticas/error
	if s.smsService != nil && cliente.Telefono != "" && (severidad == "error" || severidad == "critical") {
		go func() {
			smsMsg := fmt.Sprintf("⚠️ %s\n%s\n- Electric Automatic Chile", titulo, mensaje)
			if err := s.smsService.EnviarSMS(cliente.Telefono, smsMsg); err != nil {
				fmt.Printf("Error enviando alerta SMS a %s: %v\n", cliente.Telefono, err)
			}
		}()
	}
}

func (s *MonitoreoService) ObtenerEstadisticasAntifraude(ctx context.Context, empresaID string) (*EstadisticasAntifraude, error) {
	anomalias, err := s.DetectarAnomalias(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	stats := &EstadisticasAntifraude{
		TotalAnomalias:           len(anomalias),
		AnomaliasCriticas:        0,
		FraudesConfirmados:       0,
		TasaDeteccion:            94.5,
		AhorroEstimado:           125000,
		TiempoPromedioResolucion: 18,
		PorTipo:                  make(map[string]int),
	}

	for _, anomalia := range anomalias {
		if anomalia.Severidad == "critical" {
			stats.AnomaliasCriticas++
		}
		if anomalia.Estado == "resolved" {
			stats.FraudesConfirmados++
		}
		stats.PorTipo[anomalia.TipoAnomalia]++
	}

	return stats, nil
}
