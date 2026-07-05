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
	// Las alertas de dispositivo/monitoreo se entregan solo en la app y la
	// plataforma web: se persisten in-app y se publican en tiempo real por
	// WebSocket. NO se envían por email ni SMS (decisión de producto).
	wsNotifier *WebSocketNotifierService
	// umbralService resuelve los umbrales de alerta por empresa (con fallback
	// a los defaults del sistema si la empresa no tiene configuración propia).
	umbralService *UmbralService
	// Throttle: última alerta por dispositivo para evitar spam
	ultimaAlerta     map[string]time.Time
	throttleDuracion time.Duration
}

func NewMonitoreoService(
	notificacionRepo ports.PortNotificacion,
	dispositivoRepo ports.PortDispositivo,
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	wsNotifier *WebSocketNotifierService,
	umbralService *UmbralService,
) *MonitoreoService {
	return &MonitoreoService{
		notificacionRepo: notificacionRepo,
		dispositivoRepo:  dispositivoRepo,
		clienteRepo:      clienteRepo,
		empresaRepo:      empresaRepo,
		wsNotifier:       wsNotifier,
		umbralService:    umbralService,
		ultimaAlerta:     make(map[string]time.Time),
		throttleDuracion: 24 * time.Hour, // Máximo 1 alerta por dispositivo por día
	}
}

// resolverUmbrales obtiene los umbrales de la empresa, o los defaults si el
// servicio de umbrales no está disponible (por seguridad ante wiring parcial).
func (s *MonitoreoService) resolverUmbrales(ctx context.Context, empresaID string) Umbrales {
	if s.umbralService == nil {
		return UmbralesPorDefecto()
	}
	return s.umbralService.ObtenerUmbrales(ctx, empresaID)
}

func (s *MonitoreoService) VerificarConsumoAnormal(ctx context.Context, empresaID string) error {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return err
	}

	umbrales := s.resolverUmbrales(ctx, empresaID)

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		consumoActual := dispositivo.UltimaLectura.Energy
		
		if consumoActual > umbrales.ConsumoMax {
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
					"umbral":            umbrales.ConsumoMax,
					"numeroDispositivo": dispositivo.NumeroDispositivo,
					"dispositivoId":     dispositivo.ID.Hex(),
				},
			}

			s.notificacionRepo.Create(ctx, notificacion)
			s.publicarAlertaTiempoReal(notificacion)
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
			s.publicarAlertaTiempoReal(notificacion)
		}
	}

	return nil
}

func (s *MonitoreoService) VerificarPatronesAnormales(ctx context.Context, empresaID string) error {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return err
	}

	umbrales := s.resolverUmbrales(ctx, empresaID)

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		key := dispositivo.ID.Hex()
		if ultima, ok := s.ultimaAlerta[key]; ok && time.Since(ultima) < s.throttleDuracion {
			continue // Throttle activo — no enviar más alertas para este dispositivo
		}

		// No enviar alertas si el servicio está cortado — voltaje anormal es comportamiento esperado
		if dispositivo.EstadoServicio == "cortado" || dispositivo.EstadoServicio == "corte_pendiente" {
			continue
		}

		alertaEnviada := false

		if dispositivo.UltimaLectura.Voltage > 0 &&
			(dispositivo.UltimaLectura.Voltage < umbrales.VoltajeMin || dispositivo.UltimaLectura.Voltage > umbrales.VoltajeMax) {
			titulo := fmt.Sprintf("Voltaje anormal - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s registra %.1fV (rango normal: %.0f-%.0fV).",
				dispositivo.Nombre, dispositivo.UltimaLectura.Voltage, umbrales.VoltajeMin, umbrales.VoltajeMax)

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notif := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "error",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				FechaCreacion:  time.Now(),
			}
			s.notificacionRepo.Create(ctx, notif)
			s.publicarAlertaTiempoReal(notif)
			alertaEnviada = true
		}

		if !alertaEnviada && dispositivo.UltimaLectura.Current > umbrales.CorrienteMax {
			titulo := fmt.Sprintf("Corriente elevada - Dispositivo %s", dispositivo.NumeroDispositivo)
			mensaje := fmt.Sprintf("El dispositivo %s registra %.1fA (límite: %.0fA).",
				dispositivo.Nombre, dispositivo.UltimaLectura.Current, umbrales.CorrienteMax)

			empresaOID, _ := primitive.ObjectIDFromHex(empresaID)
			notif := &entities.NotificacionEntity{
				DestinatarioID: empresaOID,
				DispositivoID:  dispositivo.ID,
				Tipo:           "alerta",
				Severidad:      "error",
				Titulo:         titulo,
				Mensaje:        mensaje,
				Importante:     true,
				FechaCreacion:  time.Now(),
			}
			s.notificacionRepo.Create(ctx, notif)
			s.publicarAlertaTiempoReal(notif)
			alertaEnviada = true
		}

		if alertaEnviada {
			s.ultimaAlerta[key] = time.Now()
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
			empresas, err := s.empresaRepo.FindAll(ctx)
			if err != nil {
				continue
			}
			for _, empresa := range empresas {
				s.VerificarConsumoAnormal(ctx, empresa.ID)
				s.VerificarPatronesAnormales(ctx, empresa.ID)
			}
		}
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

// publicarAlertaTiempoReal empuja la notificación de dispositivo por WebSocket
// para que aparezca al instante en la app y en la plataforma web. La
// notificación ya quedó persistida in-app por el llamador.
func (s *MonitoreoService) publicarAlertaTiempoReal(notificacion *entities.NotificacionEntity) {
	if s.wsNotifier == nil || notificacion == nil {
		return
	}
	s.wsNotifier.NotificarNuevaNotificacion(notificacion)
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
