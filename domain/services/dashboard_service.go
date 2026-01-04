package services

import (
	"context"
	"electric-backend/domain/ports"
)

type DashboardService struct {
	clienteRepo     ports.PortCliente
	dispositivoRepo ports.PortDispositivo
	alertaRepo      ports.PortAlerta
	ticketRepo      ports.PortTicket
}

type EstadisticasDashboard struct {
	ClientesActivos      int     `json:"clientesActivos"`
	ClientesTotales      int     `json:"clientesTotales"`
	DispositivosActivos  int     `json:"dispositivosActivos"`
	DispositivosTotales  int     `json:"dispositivosTotales"`
	AlertasActivas       int     `json:"alertasActivas"`
	TicketsPendientes    int     `json:"ticketsPendientes"`
	ConsumoTotal         float64 `json:"consumoTotal"`
	ConsumoHoy           float64 `json:"consumoHoy"`
}

func NewDashboardService(
	clienteRepo ports.PortCliente,
	dispositivoRepo ports.PortDispositivo,
	alertaRepo ports.PortAlerta,
	ticketRepo ports.PortTicket,
) *DashboardService {
	return &DashboardService{
		clienteRepo:     clienteRepo,
		dispositivoRepo: dispositivoRepo,
		alertaRepo:      alertaRepo,
		ticketRepo:      ticketRepo,
	}
}

func (s *DashboardService) ObtenerEstadisticas(ctx context.Context, empresaID string) (*EstadisticasDashboard, error) {
	clientes, _ := s.clienteRepo.FindAll(ctx, empresaID)
	dispositivos, _ := s.dispositivoRepo.FindAll(ctx, empresaID)
	alertas, _ := s.alertaRepo.FindActivas(ctx)
	tickets, _ := s.ticketRepo.FindByEmpresa(ctx, empresaID)

	clientesActivos := 0
	for _, c := range clientes {
		if c.Activo {
			clientesActivos++
		}
	}

	dispositivosActivos := 0
	consumoTotal := 0.0
	for _, d := range dispositivos {
		if d.Activo && d.Estado == "activo" {
			dispositivosActivos++
		}
		if d.UltimaLectura != nil {
			consumoTotal += d.UltimaLectura.Energy
		}
	}

	ticketsPendientes := 0
	for _, t := range tickets {
		if t.Estado == "abierto" || t.Estado == "en_proceso" {
			ticketsPendientes++
		}
	}

	return &EstadisticasDashboard{
		ClientesActivos:      clientesActivos,
		ClientesTotales:      len(clientes),
		DispositivosActivos:  dispositivosActivos,
		DispositivosTotales:  len(dispositivos),
		AlertasActivas:       len(alertas),
		TicketsPendientes:    ticketsPendientes,
		ConsumoTotal:         consumoTotal,
		ConsumoHoy:           consumoTotal * 0.1,
	}, nil
}
