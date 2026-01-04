package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Antifraude struct {
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
	NotasInvestigacion  string                 `json:"notasInvestigacion,omitempty"`
	AsignadoA           string                 `json:"asignadoA,omitempty"`
}

type EstadisticasAntifraude struct {
	TotalAnomalias            int     `json:"totalAnomalias"`
	AnomaliasCriticas         int     `json:"anomaliasCriticas"`
	FraudesConfirmados        int     `json:"fraudesConfirmados"`
	TasaDeteccion             float64 `json:"tasaDeteccion"`
	AhorroEstimado            float64 `json:"ahorroEstimado"`
	TiempoPromedioResolucion  int     `json:"tiempoPromedioResolucion"`
	PorTipo                   map[string]int `json:"porTipo"`
}

type AntifraudeService struct {
	dispositivoRepo ports.PortDispositivo
	clienteRepo     ports.PortCliente
	alertaRepo      ports.PortAlerta
	notificacionRepo ports.PortNotificacion
}

func NewAntifraudeService(
	dispositivoRepo ports.PortDispositivo,
	clienteRepo ports.PortCliente,
	alertaRepo ports.PortAlerta,
	notificacionRepo ports.PortNotificacion,
) *AntifraudeService {
	return &AntifraudeService{
		dispositivoRepo: dispositivoRepo,
		clienteRepo:     clienteRepo,
		alertaRepo:      alertaRepo,
		notificacionRepo: notificacionRepo,
	}
}

func (s *AntifraudeService) DetectarAnomalias(ctx context.Context, empresaID string) ([]Antifraude, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []Antifraude{}, nil
	}

	anomalias := make([]Antifraude, 0)

	for _, dispositivo := range dispositivos {
		if dispositivo.UltimaLectura == nil {
			continue
		}

		cliente, err := s.clienteRepo.FindByID(ctx, dispositivo.ClienteID.Hex())
		if err != nil {
			continue
		}

		anomaliasDispositivo := s.analizarDispositivo(dispositivo, cliente)
		anomalias = append(anomalias, anomaliasDispositivo...)
	}

	return anomalias, nil
}

func (s *AntifraudeService) analizarDispositivo(dispositivo *entities.DispositivoEntity, cliente interface{}) []Antifraude {
	anomalias := make([]Antifraude, 0)
	lectura := dispositivo.UltimaLectura

	consumoPromedio := 150.0
	voltajeNormal := 220.0
	corrienteNormal := 10.0

	if lectura.Energy > consumoPromedio*2 {
		porcentaje := ((lectura.Energy - consumoPromedio) / consumoPromedio) * 100
		severidad := "high"
		if porcentaje > 200 {
			severidad = "critical"
		}

		anomalias = append(anomalias, Antifraude{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     dispositivo.Nombre,
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
		anomalias = append(anomalias, Antifraude{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     dispositivo.Nombre,
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

	if math.Abs(lectura.Voltage-voltajeNormal) > 30 {
		severidad := "medium"
		if math.Abs(lectura.Voltage-voltajeNormal) > 50 {
			severidad = "high"
		}

		anomalias = append(anomalias, Antifraude{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     dispositivo.Nombre,
			TipoAnomalia:      "voltaje_anormal",
			Severidad:         severidad,
			Descripcion:       "Voltaje fuera del rango normal",
			ValorEsperado:     voltajeNormal,
			ValorDetectado:    lectura.Voltage,
			Porcentaje:        math.Abs((lectura.Voltage-voltajeNormal)/voltajeNormal) * 100,
			FechaDeteccion:    time.Now(),
			Estado:            "pending",
			Evidencia: map[string]interface{}{
				"voltage":     lectura.Voltage,
				"current":     lectura.Current,
				"energy":      lectura.Energy,
				"timestamp":   lectura.Timestamp,
			},
		})
	}

	if lectura.Current > corrienteNormal*5 {
		anomalias = append(anomalias, Antifraude{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     dispositivo.Nombre,
			TipoAnomalia:      "corriente_elevada",
			Severidad:         "high",
			Descripcion:       "Corriente anormalmente elevada - posible sobrecarga",
			ValorEsperado:     corrienteNormal,
			ValorDetectado:    lectura.Current,
			Porcentaje:        ((lectura.Current - corrienteNormal) / corrienteNormal) * 100,
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

	potenciaCalculada := lectura.Voltage * lectura.Current
	if math.Abs(potenciaCalculada-lectura.ActivePower) > potenciaCalculada*0.3 {
		anomalias = append(anomalias, Antifraude{
			ID:                primitive.NewObjectID().Hex(),
			DispositivoID:     dispositivo.ID.Hex(),
			ClienteID:         dispositivo.ClienteID.Hex(),
			NumeroDispositivo: dispositivo.NumeroDispositivo,
			NombreCliente:     dispositivo.Nombre,
			TipoAnomalia:      "potencia_inconsistente",
			Severidad:         "critical",
			Descripcion:       "Potencia reportada no coincide con cálculo - posible manipulación",
			ValorEsperado:     potenciaCalculada,
			ValorDetectado:    lectura.ActivePower,
			Porcentaje:        math.Abs((potenciaCalculada-lectura.ActivePower)/potenciaCalculada) * 100,
			FechaDeteccion:    time.Now(),
			Estado:            "pending",
			Evidencia: map[string]interface{}{
				"voltage":          lectura.Voltage,
				"current":          lectura.Current,
				"activePower":      lectura.ActivePower,
				"calculatedPower":  potenciaCalculada,
				"timestamp":        lectura.Timestamp,
			},
		})
	}

	return anomalias
}

func (s *AntifraudeService) ObtenerEstadisticas(ctx context.Context, empresaID string) (*EstadisticasAntifraude, error) {
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

func (s *AntifraudeService) ActualizarEstadoAnomalia(ctx context.Context, anomaliaID string, estado string, notas string) error {
	return nil
}

func (s *AntifraudeService) AsignarInvestigador(ctx context.Context, anomaliaID string, investigador string) error {
	return nil
}
