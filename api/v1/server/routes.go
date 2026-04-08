package server

import (
	"electric-backend/api/v1/controllers"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/websocket"

	"github.com/gin-gonic/gin"
)

func registerRoutes(
	api *gin.RouterGroup,
	fc *facades.FacadeContainer,
	svc *services.ServiceContainer,
	repos *data.DataContainer,
	wsHub *websocket.Hub,
	arduinoBridge *arduino.SerialBridge,
) {
	// --- Controllers ---
	authCtrl := controllers.NewAuthController(fc.AuthFacade)
	clienteCtrl := controllers.NewClienteController(fc.ClienteFacade)
	empresaCtrl := controllers.NewEmpresaController(fc.EmpresaFacade)
	dispositivoCtrl := controllers.NewDispositivoController(fc.DispositivoFacade)
	notificacionCtrl := controllers.NewNotificacionController(svc.NotificacionService)
	boletaCtrl := controllers.NewBoletaController(svc.BoletaService)
	ticketCtrl := controllers.NewTicketController(svc.TicketService)
	estadisticaCtrl := controllers.NewEstadisticaController(svc.DashboardService)
	cotizacionCtrl := controllers.NewCotizacionController(fc.CotizacionFacade)
	wsCtrl := controllers.NewWebSocketController(wsHub)
	arduinoCtrl := controllers.NewArduinoController(arduinoBridge)
	dashboardClienteCtrl := controllers.NewDashboardClienteController(fc.ClienteFacade, fc.DispositivoFacade, svc.BoletaService, svc.DashboardService, arduinoBridge)
	reportesCtrl := controllers.NewReportesController(svc.ReportesService)
	mapaCtrl := controllers.NewMapaController(svc.DispositivoService, svc.ClienteService)
	antifraudeCtrl := controllers.NewAntifraudeController(svc.MonitoreoService)
	dashboardCtrl := controllers.NewDashboardController(svc.DashboardService)
	tarifaCtrl := controllers.NewTarifaController(svc.TarifaService)
	consumoCtrl := controllers.NewConsumoController(svc.ConsumoService)
	iaCtrl := controllers.NewIAController(svc.IAService)
	historialConsumoCtrl := controllers.NewHistorialConsumoController()
	usuarioEmpresaCtrl := controllers.NewUsuarioEmpresaController(svc.UsuarioEmpresaService)
	iotStatusCtrl := controllers.NewIoTStatusController(repos.DispositivoRepo)

	// --- Registrar todas las rutas via SetupRoutes ---
	authCtrl.SetupRoutes(api)
	clienteCtrl.SetupRoutes(api)
	dashboardClienteCtrl.SetupRoutes(api)
	cotizacionCtrl.SetupRoutes(api)
	historialConsumoCtrl.SetupRoutes(api)
	iaCtrl.SetupRoutes(api)
	empresaCtrl.SetupRoutes(api)
	dispositivoCtrl.SetupRoutes(api)
	notificacionCtrl.SetupRoutes(api)
	boletaCtrl.SetupRoutes(api)
	ticketCtrl.SetupRoutes(api)
	estadisticaCtrl.SetupRoutes(api)
	reportesCtrl.SetupRoutes(api)
	mapaCtrl.SetupRoutes(api)
	antifraudeCtrl.SetupRoutes(api)
	dashboardCtrl.SetupRoutes(api)
	tarifaCtrl.SetupRoutes(api)
	consumoCtrl.SetupRoutes(api)
	usuarioEmpresaCtrl.SetupRoutes(api)
	wsCtrl.SetupRoutes(api)
	arduinoCtrl.SetupRoutes(api)
	iotStatusCtrl.SetupRoutes(api)

	// --- IoT (API Key) — rutas separadas con IoTAPIKeyMiddleware ---
	dispositivoCtrl.SetupIoTRoutes(api)

	// --- Imágenes de perfil (solo si S3 disponible) ---
	if svc.ImagenPerfilService != nil {
		imgCtrl := controllers.NewImagenPerfilController(svc.ImagenPerfilService)
		imgCtrl.SetupRoutes(api)
	}
}
