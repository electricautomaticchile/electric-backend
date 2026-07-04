package services

import (
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/eventbus"
	"electric-backend/infrastructure/push"
	"electric-backend/infrastructure/sms"
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
	FCMTokenService        *FCMTokenService
	PushService            push.PushService
}

// ExternalDeps groups external service dependencies.
type ExternalDeps struct {
	WSPublisher *eventbus.Publisher
	EmailSvc    email.EmailService
	SMSSvc      sms.SMSService
	PushSvc     push.PushService
}

func Build(repos *data.DataContainer, ext *ExternalDeps) *ServiceContainer {
	wsNotifier := NewWebSocketNotifierService(ext.WSPublisher)

	authService := NewAuthService(repos.EmpresaRepo, repos.ClienteRepo, repos.UsuarioEmpresaRepo, repos.RecoveryTokenRepo, repos.RefreshTokenRepo, ext.EmailSvc)
	clienteService := NewClienteService(repos.ClienteRepo, ext.EmailSvc)
	empresaService := NewEmpresaService(repos.EmpresaRepo)
	dispositivoService := NewDispositivoService(repos.DispositivoRepo, repos.ClienteRepo, wsNotifier)
	notificacionService := NewNotificacionService(repos.NotificacionRepo, wsNotifier)
	boletaService := NewBoletaService(repos.BoletaRepo, repos.ClienteRepo, ext.EmailSvc)
	ticketService := NewTicketService(repos.TicketRepo, repos.NotificacionRepo, ext.EmailSvc, repos.ClienteRepo, repos.EmpresaRepo)
	cotizacionService := NewCotizacionService(repos.CotizacionRepo)
	monitoreoService := NewMonitoreoService(repos.NotificacionRepo, repos.DispositivoRepo, repos.ClienteRepo, repos.EmpresaRepo, ext.EmailSvc, ext.SMSSvc)
	tarifaService := NewTarifaService(repos.TarifaRepo)
	consumoService := NewConsumoService(repos.ClienteRepo, repos.TarifaRepo)
	dashboardService := NewDashboardService(repos.ClienteRepo, repos.DispositivoRepo, repos.NotificacionRepo, repos.TicketRepo, repos.EstadisticaRepo)
	reportesService := NewReportesService(repos.ClienteRepo, repos.DispositivoRepo, repos.NotificacionRepo, repos.BoletaRepo)
	usuarioEmpresaService := NewUsuarioEmpresaService(repos.UsuarioEmpresaRepo, ext.EmailSvc)
	notificacionSMSService := NewNotificacionSMSService(repos.ClienteRepo, repos.BoletaRepo, repos.DispositivoRepo, ext.SMSSvc)
	auditLogService := NewAuditLogService(repos.AuditLogRepo)
	fcmTokenService := NewFCMTokenService(repos.FCMTokenRepo)

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
		ImagenPerfilService:    nil,
		NotificacionSMSService: notificacionSMSService,
		AuditLogService:        auditLogService,
		IAService:              NewIAService(),
		FCMTokenService:        fcmTokenService,
		PushService:            ext.PushSvc,
	}
}
