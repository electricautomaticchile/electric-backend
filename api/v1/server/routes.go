package server

import (
	"electric-backend/api/v1/controllers"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/middleware"
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

	// --- Rutas con SetupRoutes propio ---
	authCtrl.SetupRoutes(api)
	dashboardClienteCtrl.SetupRoutes(api)
	clienteCtrl.SetupRoutes(api)
	cotizacionCtrl.SetupRoutes(api)
	historialConsumoCtrl.SetupRoutes(api)
	iaCtrl.SetupRoutes(api)

	// --- Imágenes de perfil (solo si S3 disponible) ---
	if svc.ImagenPerfilService != nil {
		imgCtrl := controllers.NewImagenPerfilController(svc.ImagenPerfilService)
		img := api.Group("/imagenes-perfil")
		img.Use(middleware.AuthMiddleware())
		img.POST("/:tipoUsuario/:userId/upload", imgCtrl.SubirYActualizarImagen)
		img.GET("/:tipoUsuario/:userId", imgCtrl.ObtenerImagenPerfil)
		img.DELETE("/:tipoUsuario/:userId", imgCtrl.EliminarImagenPerfil)
	}

	// --- Usuarios empresa (con CSRF) ---
	ue := api.Group("/usuarios-empresa")
	ue.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	ue.GET("", usuarioEmpresaCtrl.ObtenerTodos)
	ue.GET("/:id", usuarioEmpresaCtrl.ObtenerPorID)
	ue.POST("", usuarioEmpresaCtrl.Crear)
	ue.PUT("/:id", usuarioEmpresaCtrl.Actualizar)
	ue.DELETE("/:id", usuarioEmpresaCtrl.Eliminar)

	// --- WebSocket ---
	ws := api.Group("/ws")
	ws.GET("/connect", wsCtrl.HandleWebSocket)
	ws.GET("/stats", wsCtrl.GetStats)

	// --- Arduino ---
	ard := api.Group("/arduino")
	ard.Use(middleware.AuthMiddleware())
	ard.GET("/status", arduinoCtrl.GetStatus)
	ard.GET("/ports", arduinoCtrl.ListPorts)
	ard.POST("/connect", arduinoCtrl.Connect)
	ard.POST("/disconnect", arduinoCtrl.Disconnect)
	ard.POST("/command", arduinoCtrl.SendCommand)

	// --- CRUD con auth ---
	registerCRUDRoutes(api, empresaCtrl, dispositivoCtrl, notificacionCtrl, boletaCtrl, ticketCtrl)
	registerDashboardRoutes(api, estadisticaCtrl, reportesCtrl, mapaCtrl, antifraudeCtrl, dashboardCtrl, tarifaCtrl, consumoCtrl)

	// --- IoT (API Key) ---
	iot := api.Group("/iot")
	iot.Use(middleware.IoTAPIKeyMiddleware())
	iot.POST("/lectura", dispositivoCtrl.RecibirLecturaIoT)
	iot.POST("/comando-ejecutado", dispositivoCtrl.ComandoEjecutado)

	// --- IoT Status (JWT) ---
	iotV1 := api.Group("/v1/iot")
	iotV1.Use(middleware.AuthMiddleware())
	iotV1.GET("/status", iotStatusCtrl.GetAllStatus)
}

func registerCRUDRoutes(api *gin.RouterGroup, empresa *controllers.EmpresaController, dispositivo *controllers.DispositivoController, notificacion *controllers.NotificacionController, boleta *controllers.BoletaController, ticket *controllers.TicketController) {
	e := api.Group("/empresas")
	e.Use(middleware.AuthMiddleware())
	e.GET("", empresa.ObtenerTodas)
	e.GET("/:id", empresa.ObtenerPorID)
	e.POST("", empresa.Crear)
	e.PUT("/:id", empresa.Actualizar)
	e.DELETE("/:id", empresa.Eliminar)

	d := api.Group("/dispositivos")
	d.Use(middleware.AuthMiddleware())
	d.GET("", dispositivo.ObtenerTodos)
	d.GET("/:id", dispositivo.ObtenerPorID)
	d.POST("", dispositivo.Crear)
	d.PUT("/:id", dispositivo.Actualizar)
	d.PUT("/:id/asignar", dispositivo.AsignarCliente)
	d.PUT("/:id/desasignar", dispositivo.DesasignarCliente)
	d.DELETE("/:id", dispositivo.Eliminar)

	n := api.Group("/notificaciones")
	n.Use(middleware.AuthMiddleware())
	n.GET("", notificacion.Listar)
	n.PUT("/:id/marcar-leida", notificacion.MarcarLeida)
	n.PUT("/marcar-todas-leidas", notificacion.MarcarTodasLeidas)
	n.DELETE("/:id", notificacion.Eliminar)
	n.GET("/estadisticas", notificacion.ObtenerEstadisticas)

	b := api.Group("/boletas")
	b.Use(middleware.AuthMiddleware())
	b.GET("/cliente/:clienteId", boleta.ObtenerPorCliente)
	b.POST("", boleta.Crear)

	t := api.Group("/tickets")
	t.Use(middleware.AuthMiddleware())
	t.GET("", ticket.ObtenerTodos)
	t.GET("/:id", ticket.ObtenerPorID)
	t.GET("/cliente/:clienteId", ticket.ObtenerPorCliente)
	t.GET("/empresa/:empresaId", ticket.ObtenerPorEmpresa)
	t.POST("", ticket.Crear)
	t.PUT("/:id/responder", ticket.AgregarRespuesta)
	t.PUT("/:id/estado", ticket.ActualizarEstado)
	t.DELETE("/:id", ticket.Eliminar)
}

func registerDashboardRoutes(api *gin.RouterGroup, estadistica *controllers.EstadisticaController, reportes *controllers.ReportesController, mapa *controllers.MapaController, antifraude *controllers.AntifraudeController, dashboard *controllers.DashboardController, tarifa *controllers.TarifaController, consumo *controllers.ConsumoController) {
	est := api.Group("/estadisticas")
	est.Use(middleware.AuthMiddleware())
	est.GET("/cliente/:clienteId", estadistica.ObtenerConsumoCliente)
	est.GET("/globales", estadistica.ObtenerEstadisticasGlobales)
	est.GET("/resumen", estadistica.ObtenerResumen)

	rep := api.Group("/reportes")
	rep.Use(middleware.AuthMiddleware())
	rep.GET("/clientes", reportes.Clientes)
	rep.GET("/dispositivos", reportes.Dispositivos)
	rep.GET("/alertas", reportes.Alertas)
	rep.GET("/boletas", reportes.Boletas)
	rep.GET("/consumo", reportes.Consumo)

	m := api.Group("/mapa")
	m.Use(middleware.AuthMiddleware())
	m.GET("/datos", mapa.ObtenerDatosMapa)

	af := api.Group("/antifraude")
	af.Use(middleware.AuthMiddleware())
	af.GET("/anomalias", antifraude.DetectarAnomalias)
	af.GET("/estadisticas", antifraude.ObtenerEstadisticas)

	db := api.Group("/dashboard")
	db.Use(middleware.AuthMiddleware())
	db.GET("/estadisticas", dashboard.ObtenerEstadisticas)

	tf := api.Group("/tarifas")
	tf.Use(middleware.AuthMiddleware())
	tf.GET("", tarifa.ObtenerTodas)
	tf.GET("/activa", tarifa.ObtenerActiva)
	tf.GET("/:id", tarifa.ObtenerPorID)
	tf.POST("", tarifa.Crear)
	tf.PUT("/:id", tarifa.Actualizar)
	tf.DELETE("/:id", tarifa.Eliminar)
	tf.POST("/calcular", tarifa.CalcularConsumo)

	c := api.Group("/consumo")
	c.Use(middleware.AuthMiddleware())
	c.GET("/cliente/:clienteId/calcular", consumo.CalcularCostoActual)
	c.GET("/cliente/:clienteId/actual", consumo.ObtenerConsumoActual)
}
