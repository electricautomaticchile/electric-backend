package services

import (
	"electric-backend/infrastructure/aws"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/sms"
	"electric-backend/infrastructure/websocket"
)

// ServiceContainer agrupa todos los servicios de dominio.
type ServiceContainer struct {
	AuthService            *AuthService
	ClienteService         *ClienteService
	EmpresaService         *EmpresaService
	DispositivoService     *DispositivoService
	NotificacionService    *NotificacionService
	BoletaService          *BoletaService
	TicketService          *TicketService
	CotizacionService      *CotizacionService
	MonitoreoService       *MonitoreoService
	TarifaService          *TarifaService
	ConsumoService         *ConsumoService
	DashboardService       *DashboardService
	ReportesService        *ReportesService
	UsuarioEmpresaService  *UsuarioEmpresaService
	ImagenPerfilService    *ImagenPerfilService
	NotificacionSMSService *NotificacionSMSService
	AuditLogService        *AuditLogService
	IAService              *IAService
}

func Build(repos *data.DataContainer, wsHub *websocket.Hub, emailSvc email.EmailService, smsSvc sms.SMSService, s3Svc *aws.S3Service) *ServiceContainer {
	wsNotifier := NewWebSocketNotifierService(wsHub)

	authService := NewAuthService(repos.EmpresaRepo, repos.ClienteRepo, repos.UsuarioEmpresaRepo, repos.RecoveryTokenRepo, repos.RefreshTokenRepo, emailSvc)
	clienteService := NewClienteService(repos.ClienteRepo, emailSvc)
	empresaService := NewEmpresaService(repos.EmpresaRepo)
	dispositivoService := NewDispositivoService(repos.DispositivoRepo, repos.ClienteRepo, wsNotifier)
	notificacionService := NewNotificacionService(repos.NotificacionRepo, wsNotifier)
	boletaService := NewBoletaService(repos.BoletaRepo, repos.ClienteRepo, emailSvc)
	ticketService := NewTicketService(repos.TicketRepo, repos.NotificacionRepo, emailSvc, repos.ClienteRepo, repos.EmpresaRepo)
	cotizacionService := NewCotizacionService(repos.CotizacionRepo)
	monitoreoService := NewMonitoreoService(repos.NotificacionRepo, repos.DispositivoRepo, repos.ClienteRepo, repos.EmpresaRepo, emailSvc, smsSvc)
	tarifaService := NewTarifaService(repos.TarifaRepo)
	consumoService := NewConsumoService(repos.ClienteRepo, repos.TarifaRepo)
	dashboardService := NewDashboardService(repos.ClienteRepo, repos.DispositivoRepo, repos.NotificacionRepo, repos.TicketRepo, repos.EstadisticaRepo)
	reportesService := NewReportesService(repos.ClienteRepo, repos.DispositivoRepo, repos.NotificacionRepo, repos.BoletaRepo)
	usuarioEmpresaService := NewUsuarioEmpresaService(repos.UsuarioEmpresaRepo, emailSvc)
	notificacionSMSService := NewNotificacionSMSService(repos.ClienteRepo, repos.BoletaRepo, repos.DispositivoRepo, smsSvc)
	auditLogService := NewAuditLogService(repos.AuditLogRepo)

	var imagenPerfilService *ImagenPerfilService
	if s3Svc != nil {
		imagenPerfilService = NewImagenPerfilService(repos.ClienteRepo, repos.EmpresaRepo, s3Svc)
	}

	return &ServiceContainer{
		AuthService:            authService,
		ClienteService:         clienteService,
		EmpresaService:         empresaService,
		DispositivoService:     dispositivoService,
		NotificacionService:    notificacionService,
		BoletaService:          boletaService,
		TicketService:          ticketService,
		CotizacionService:      cotizacionService,
		MonitoreoService:       monitoreoService,
		TarifaService:          tarifaService,
		ConsumoService:         consumoService,
		DashboardService:       dashboardService,
		ReportesService:        reportesService,
		UsuarioEmpresaService:  usuarioEmpresaService,
		ImagenPerfilService:    imagenPerfilService,
		NotificacionSMSService: notificacionSMSService,
		AuditLogService:        auditLogService,
		IAService:              NewIAService(),
	}
}
