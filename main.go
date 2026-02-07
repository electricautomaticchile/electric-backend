package main

import (
	"context"
	"electric-backend/api/v1/controllers"
	"electric-backend/config"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/aws"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/scheduler"
	"electric-backend/infrastructure/validation"
	"electric-backend/infrastructure/websocket"
	"electric-backend/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func main() {
	config.LoadConfig()

	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("rut", func(fl validator.FieldLevel) bool {
			return validation.ValidarRUT(fl.Field().String())
		})
		v.RegisterValidation("password_strong", func(fl validator.FieldLevel) bool {
			return validation.ValidarPassword(fl.Field().String()) == nil
		})
		v.RegisterValidation("telefono_cl", func(fl validator.FieldLevel) bool {
			return validation.ValidarTelefonoChileno(fl.Field().String())
		})
		v.RegisterValidation("periodo", func(fl validator.FieldLevel) bool {
			return validation.ValidarPeriodo(fl.Field().String())
		})
		v.RegisterValidation("numero_cliente", func(fl validator.FieldLevel) bool {
			return validation.ValidarNumeroCliente(fl.Field().String())
		})
	}

	if err := config.ConnectDatabase(config.AppConfig.MongoURI); err != nil {
		log.Fatalf("❌ Error conectando a MongoDB: %v", err)
	}
	defer config.DisconnectDatabase()

	if err := config.CreateIndexes(config.MongoDB); err != nil {
		log.Printf("⚠️ Error creando índices: %v", err)
	}

	if err := config.ConnectRedis(
		config.AppConfig.RedisHost,
		config.AppConfig.RedisPort,
		config.AppConfig.RedisPassword,
		config.AppConfig.RedisDB,
	); err != nil {
		log.Printf("⚠️ Redis no disponible: %v", err)
	}
	defer config.DisconnectRedis()

	auditLogRepo := data.NewAuditLogRepository()
	auditLogService := services.NewAuditLogService(auditLogRepo)

	router := gin.Default()

	router.Use(middleware.CompressionMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.AuditMiddleware(auditLogService))
	router.Use(middleware.ErrorHandler())

	rateLimits := map[string]int{
		"POST:/api/auth/login":                    5,
		"POST:/api/auth/register":                 3,
		"POST:/api/auth/recovery":                 3,
		"POST:/api/auth/reset-password":           3,
		"POST:/api/auth/cambiar-password":         10,
		"POST:/api/auth/refresh-token":            20,
		
		"GET:/api/clientes":                       60,
		"POST:/api/clientes":                      20,
		"PUT:/api/clientes/:id":                   30,
		"DELETE:/api/clientes/:id":                10,
		"GET:/api/clientes/:id":                   100,
		
		"GET:/api/dispositivos":                   120,
		"POST:/api/dispositivos":                  20,
		"PUT:/api/dispositivos/:id":               60,
		"POST:/api/dispositivos/lectura":          300,
		"POST:/api/dispositivos/control":          30,
		
		"GET:/api/alertas":                        100,
		"POST:/api/alertas":                       50,
		"PUT:/api/alertas/:id":                    40,
		"DELETE:/api/alertas/:id":                 20,
		
		"GET:/api/boletas":                        60,
		"POST:/api/boletas":                       10,
		"GET:/api/boletas/:id":                    100,
		
		"GET:/api/tickets":                        60,
		"POST:/api/tickets":                       10,
		"POST:/api/tickets/:id/responder":         20,
		"PUT:/api/tickets/:id":                    30,
		
		"GET:/api/dashboard/empresa":              120,
		"GET:/api/dashboard/cliente":              120,
		"GET:/api/estadisticas/*":                 100,
		
		"GET:/api/export/*":                       5,
		
		"GET:/api/mapa/dispositivos":              60,
		"GET:/api/mapa/ubicacion/:id":             100,
		"POST:/api/mapa/actualizar":               30,
		
		"GET:/api/tarifas":                        30,
		"POST:/api/tarifas":                       5,
		"PUT:/api/tarifas/:id":                    10,
		
		"POST:/api/imagenes-perfil/upload":        10,
		"GET:/api/imagenes-perfil/:id":            100,
		"DELETE:/api/imagenes-perfil/:id":         10,
	}

	router.Use(middleware.EndpointRateLimitMiddleware(rateLimits))

	wsHub := websocket.InitializeHub()
	
	go wsHub.Run()
	
	arduinoBridge := arduino.NewSerialBridge(wsHub)
	
	if os.Getenv("ARDUINO_ENABLED") == "true" {
		go func() {
			time.Sleep(2 * time.Second)
			if err := arduinoBridge.Connect(""); err != nil {
				log.Printf("⚠️ No se pudo conectar al Arduino: %v", err)
				log.Printf("💡 El servidor continuará sin Arduino. Para conectar: reinicia con Arduino conectado")
			}
		}()
	} else {
		log.Printf("ℹ️ Arduino deshabilitado. Para habilitar: ARDUINO_ENABLED=true")
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"message":     "API Electricautomaticchile funcionando correctamente",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "2.0.0",
			"environment": config.AppConfig.Environment,
			"database": gin.H{
				"connected": config.MongoDB != nil,
			},
			"redis": gin.H{
				"connected": config.RedisClient != nil,
			},
			"websocket": gin.H{
				"clients": wsHub.GetConnectedClients(),
			},
			"arduino": gin.H{
				"connected": arduinoBridge.IsConnected(),
				"devices":   len(arduinoBridge.GetDevices()),
			},
		})
	})

	// Inicializar repositorios
	recoveryTokenRepo := data.NewRecoveryTokenRepository()
	clienteRepo := data.NewClienteRepository()
	empresaRepo := data.NewEmpresaRepository()
	dispositivoRepo := data.NewDispositivoRepository()
	notificacionRepo := data.NewNotificacionRepository()
	boletaRepo := data.NewBoletaRepository()
	ticketRepo := data.NewTicketRepository()
	configuracionRepo := data.NewConfiguracionRepository()
	estadisticaRepo := data.NewEstadisticaRepository()
	cotizacionRepo := data.NewCotizacionRepository()
	tarifaRepo := data.NewTarifaRepository()

	wsNotifierService := services.NewWebSocketNotifierService(wsHub)
	emailService := email.NewResendService()
	dashboardService := services.NewDashboardService(clienteRepo, dispositivoRepo, notificacionRepo, ticketRepo, estadisticaRepo)
	
	s3Service, err := aws.NewS3Service(config.AppConfig)
	if err != nil {
		log.Printf("⚠️ S3 no disponible: %v", err)
		s3Service = nil
	} else {
		log.Printf("✅ S3 inicializado correctamente")
	}
	
	var imagenPerfilService *services.ImagenPerfilService
	if s3Service != nil {
		imagenPerfilService = services.NewImagenPerfilService(clienteRepo, empresaRepo, s3Service)
		log.Printf("✅ Servicio de imágenes de perfil habilitado")
	} else {
		log.Printf("⚠️ Servicio de imágenes de perfil deshabilitado (S3 no configurado)")
	}
	
	exportService := services.NewExportService(clienteRepo, dispositivoRepo, notificacionRepo, boletaRepo)

	usuarioEmpresaRepo := data.NewUsuarioEmpresaRepository()
	usuarioEmpresaService := services.NewUsuarioEmpresaService(usuarioEmpresaRepo, emailService)

	refreshTokenRepo := data.NewRefreshTokenRepository()
	authService := services.NewAuthService(empresaRepo, clienteRepo, usuarioEmpresaRepo, recoveryTokenRepo, refreshTokenRepo, emailService)
	clienteService := services.NewClienteService(clienteRepo, emailService)
	empresaService := services.NewEmpresaService(empresaRepo)
	dispositivoService := services.NewDispositivoService(dispositivoRepo, clienteRepo, wsNotifierService)
	notificacionService := services.NewNotificacionService(notificacionRepo, wsNotifierService)
	boletaService := services.NewBoletaService(boletaRepo, clienteRepo, emailService)
	ticketService := services.NewTicketService(ticketRepo, notificacionRepo, emailService, clienteRepo, empresaRepo)
	configuracionService := services.NewConfiguracionService(configuracionRepo)
	cotizacionService := services.NewCotizacionService(cotizacionRepo)
	alertaAutomaticaService := services.NewMonitoreoService(notificacionRepo, dispositivoRepo, clienteRepo, empresaRepo)
	reporteService := services.NewReporteService(clienteRepo, empresaRepo, dispositivoRepo, boletaRepo)
	monitoreoService := alertaAutomaticaService
	tarifaService := services.NewTarifaService(tarifaRepo)
	consumoService := services.NewConsumoService(clienteRepo, tarifaRepo)

	notificacionSMSService := services.NewNotificacionSMSService(clienteRepo, boletaRepo, dispositivoRepo)
	notificacionesScheduler := scheduler.NewNotificacionesScheduler(notificacionSMSService)
	notificacionesScheduler.Iniciar()
	log.Printf("✅ Scheduler de notificaciones SMS iniciado")

	notificacionSMSController := controllers.NewNotificacionSMSController(notificacionSMSService)

	// Inicializar facades
	authFacade := facades.NewAuthFacade(authService)
	clienteFacade := facades.NewClienteFacade(clienteService)
	empresaFacade := facades.NewEmpresaFacade(empresaService)
	dispositivoFacade := facades.NewDispositivoFacade(dispositivoService)
	cotizacionFacade := facades.NewCotizacionFacade(cotizacionService)

	iaService := services.NewIAService()

	authController := controllers.NewAuthController(authFacade)
	usuarioEmpresaController := controllers.NewUsuarioEmpresaController(usuarioEmpresaService)
	clienteController := controllers.NewClienteController(clienteFacade)
	empresaController := controllers.NewEmpresaController(empresaFacade)
	dispositivoController := controllers.NewDispositivoController(dispositivoFacade)
	notificacionController := controllers.NewNotificacionController(notificacionService)
	boletaController := controllers.NewBoletaController(boletaService)
	ticketController := controllers.NewTicketController(ticketService)
	configuracionController := controllers.NewConfiguracionController(configuracionService)
	estadisticaController := controllers.NewEstadisticaController(dashboardService)
	cotizacionController := controllers.NewCotizacionController(cotizacionFacade)
	wsController := controllers.NewWebSocketController(wsHub)
	arduinoController := controllers.NewArduinoController(arduinoBridge)
	dashboardClienteController := controllers.NewDashboardClienteController(clienteFacade, dispositivoFacade, boletaService, dashboardService)
	reporteController := controllers.NewReporteController(reporteService)
	mapaController := controllers.NewMapaController(dispositivoService, clienteService)
	antifraudeController := controllers.NewAntifraudeController(monitoreoService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	exportController := controllers.NewExportController(exportService)
	auditLogController := controllers.NewAuditLogController(auditLogService)
	tarifaController := controllers.NewTarifaController(tarifaService)
	consumoController := controllers.NewConsumoController(consumoService)
	iaController := controllers.NewIAController(iaService)

	api := router.Group("/api")
	{
		authController.SetupRoutes(api)
		clienteController.SetupRoutes(api)
		cotizacionController.SetupRoutes(api)
		dashboardClienteController.SetupRoutes(api)
		iaController.SetupRoutes(api)
		
		if imagenPerfilService != nil {
			imagenPerfilController := controllers.NewImagenPerfilController(imagenPerfilService)
			setupImagenPerfilRoutes(api, imagenPerfilController)
		}

		usuariosEmpresa := api.Group("/usuarios-empresa")
		usuariosEmpresa.Use(middleware.AuthMiddleware())
		{
			usuariosEmpresa.GET("", usuarioEmpresaController.ObtenerTodos)
			usuariosEmpresa.GET("/:id", usuarioEmpresaController.ObtenerPorID)
			usuariosEmpresa.POST("", usuarioEmpresaController.Crear)
			usuariosEmpresa.PUT("/:id", usuarioEmpresaController.Actualizar)
			usuariosEmpresa.DELETE("/:id", usuarioEmpresaController.Eliminar)
		}

		ws := api.Group("/ws")
		ws.Use(middleware.AuthMiddleware())
		{
			ws.GET("/connect", wsController.HandleWebSocket)
			ws.GET("/stats", wsController.GetStats)
		}

		arduino := api.Group("/arduino")
		arduino.Use(middleware.AuthMiddleware())
		{
			arduino.GET("/status", arduinoController.GetStatus)
			arduino.GET("/ports", arduinoController.ListPorts)
			arduino.POST("/connect", arduinoController.Connect)
			arduino.POST("/disconnect", arduinoController.Disconnect)
			arduino.POST("/command", arduinoController.SendCommand)
		}

		setupEmpresaRoutes(api, empresaController)
		setupDispositivoRoutes(api, dispositivoController)
		setupNotificacionRoutes(api, notificacionController)
		setupNotificacionSMSRoutes(api, notificacionSMSController)
		setupBoletaRoutes(api, boletaController)
		setupTicketRoutes(api, ticketController)
		setupConfiguracionRoutes(api, configuracionController)
		setupEstadisticaRoutes(api, estadisticaController)
		setupReporteRoutes(api, reporteController)
		setupMapaRoutes(api, mapaController)
		setupAntifraudeRoutes(api, antifraudeController)
		setupDashboardRoutes(api, dashboardController)
		setupExportRoutes(api, exportController)
		setupAuditLogRoutes(api, auditLogController)
		setupTarifaRoutes(api, tarifaController)
		setupConsumoRoutes(api, consumoController)
	}

	ctx := context.Background()
	go alertaAutomaticaService.IniciarMonitoreoAutomatico(ctx)

	// Ruta 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.ApiResponse{
			Success: false,
			Error:   "Ruta no encontrada",
		})
	})

	// Iniciar servidor
	port := config.AppConfig.Port
	log.Printf("✅ Servidor corriendo en puerto %s", port)
	log.Printf("🌍 Entorno: %s", config.AppConfig.Environment)
	log.Printf("📡 API disponible en: http://localhost:%s/api", port)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error iniciando servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Apagando servidor...")
	
	notificacionesScheduler.Detener()
	log.Println("✅ Scheduler de notificaciones detenido")
	
	arduinoBridge.Disconnect()
	log.Println("✅ Arduino desconectado")
}

// Funciones temporales de rutas (mover a controladores después)
func setupEmpresaRoutes(router *gin.RouterGroup, ctrl *controllers.EmpresaController) {
	empresas := router.Group("/empresas")
	empresas.Use(middleware.AuthMiddleware())
	{
		empresas.GET("", ctrl.ObtenerTodas)
		empresas.GET("/:id", ctrl.ObtenerPorID)
		empresas.POST("", ctrl.Crear)
		empresas.PUT("/:id", ctrl.Actualizar)
		empresas.DELETE("/:id", ctrl.Eliminar)
	}
}

func setupDispositivoRoutes(router *gin.RouterGroup, ctrl *controllers.DispositivoController) {
	dispositivos := router.Group("/dispositivos")
	dispositivos.Use(middleware.AuthMiddleware())
	{
		dispositivos.GET("", ctrl.ObtenerTodos)
		dispositivos.GET("/:id", ctrl.ObtenerPorID)
		dispositivos.POST("", ctrl.Crear)
		dispositivos.PUT("/:id", ctrl.Actualizar)
		dispositivos.PUT("/:id/asignar", ctrl.AsignarCliente)
		dispositivos.PUT("/:id/desasignar", ctrl.DesasignarCliente)
		dispositivos.DELETE("/:id", ctrl.Eliminar)
	}
}

func setupNotificacionRoutes(router *gin.RouterGroup, ctrl *controllers.NotificacionController) {
	notificaciones := router.Group("/notificaciones")
	notificaciones.Use(middleware.AuthMiddleware())
	{
		notificaciones.GET("", ctrl.Listar)
		notificaciones.PUT("/:id/marcar-leida", ctrl.MarcarLeida)
		notificaciones.PUT("/marcar-todas-leidas", ctrl.MarcarTodasLeidas)
		notificaciones.DELETE("/:id", ctrl.Eliminar)
		notificaciones.GET("/estadisticas", ctrl.ObtenerEstadisticas)
	}
}

func setupNotificacionSMSRoutes(router *gin.RouterGroup, ctrl *controllers.NotificacionSMSController) {
	sms := router.Group("/notificaciones-sms")
	sms.Use(middleware.AuthMiddleware())
	{
		sms.POST("/enviar-manual", ctrl.EnviarNotificacionManual)
		sms.POST("/verificar-boletas", ctrl.VerificarBoletasImpagas)
	}
}

func setupBoletaRoutes(router *gin.RouterGroup, ctrl *controllers.BoletaController) {
	boletas := router.Group("/boletas")
	boletas.Use(middleware.AuthMiddleware())
	{
		boletas.GET("/cliente/:clienteId", ctrl.ObtenerPorCliente)
		boletas.POST("", ctrl.Crear)
	}
}

func setupTicketRoutes(router *gin.RouterGroup, ctrl *controllers.TicketController) {
	tickets := router.Group("/tickets")
	tickets.Use(middleware.AuthMiddleware())
	{
		tickets.GET("", ctrl.ObtenerTodos)
		tickets.GET("/:id", ctrl.ObtenerPorID)
		tickets.GET("/cliente/:clienteId", ctrl.ObtenerPorCliente)
		tickets.GET("/empresa/:empresaId", ctrl.ObtenerPorEmpresa)
		tickets.POST("", ctrl.Crear)
		tickets.PUT("/:id/responder", ctrl.AgregarRespuesta)
		tickets.PUT("/:id/estado", ctrl.ActualizarEstado)
		tickets.DELETE("/:id", ctrl.Eliminar)
	}
}

func setupConfiguracionRoutes(router *gin.RouterGroup, ctrl *controllers.ConfiguracionController) {
	configuracion := router.Group("/configuracion")
	configuracion.Use(middleware.AuthMiddleware())
	{
		configuracion.GET("", ctrl.ObtenerTodas)
		configuracion.GET("/:clave", ctrl.ObtenerPorClave)
		configuracion.PUT("", ctrl.Actualizar)
	}
}

func setupEstadisticaRoutes(router *gin.RouterGroup, ctrl *controllers.EstadisticaController) {
	estadisticas := router.Group("/estadisticas")
	estadisticas.Use(middleware.AuthMiddleware())
	{
		estadisticas.GET("/cliente/:clienteId", ctrl.ObtenerConsumoCliente)
		estadisticas.GET("/globales", ctrl.ObtenerEstadisticasGlobales)
	}
}

func setupReporteRoutes(router *gin.RouterGroup, ctrl *controllers.ReporteController) {
	reportes := router.Group("/reportes")
	reportes.Use(middleware.AuthMiddleware())
	{
		reportes.GET("/clientes", ctrl.GenerarReporteClientes)
		reportes.GET("/dispositivos", ctrl.GenerarReporteDispositivos)
		reportes.GET("/boletas", ctrl.GenerarReporteBoletas)
		reportes.GET("/consumo", ctrl.GenerarReporteConsumo)
	}
}

func setupMapaRoutes(router *gin.RouterGroup, ctrl *controllers.MapaController) {
	mapa := router.Group("/mapa")
	mapa.Use(middleware.AuthMiddleware())
	{
		mapa.GET("/datos", ctrl.ObtenerDatosMapa)
	}
}

func setupAntifraudeRoutes(router *gin.RouterGroup, ctrl *controllers.AntifraudeController) {
	antifraude := router.Group("/antifraude")
	antifraude.Use(middleware.AuthMiddleware())
	{
		antifraude.GET("/anomalias", ctrl.DetectarAnomalias)
		antifraude.GET("/estadisticas", ctrl.ObtenerEstadisticas)
	}
}

func setupDashboardRoutes(router *gin.RouterGroup, ctrl *controllers.DashboardController) {
	dashboard := router.Group("/dashboard")
	dashboard.Use(middleware.AuthMiddleware())
	{
		dashboard.GET("/estadisticas", ctrl.ObtenerEstadisticas)
	}
}

func setupImagenPerfilRoutes(router *gin.RouterGroup, ctrl *controllers.ImagenPerfilController) {
	imagenes := router.Group("/imagenes-perfil")
	imagenes.Use(middleware.AuthMiddleware())
	{
		imagenes.POST("/:tipoUsuario/:userId/upload", ctrl.SubirYActualizarImagen)
		imagenes.GET("/:tipoUsuario/:userId", ctrl.ObtenerImagenPerfil)
		imagenes.DELETE("/:tipoUsuario/:userId", ctrl.EliminarImagenPerfil)
	}
}

func setupExportRoutes(router *gin.RouterGroup, ctrl *controllers.ExportController) {
	export := router.Group("/export")
	export.Use(middleware.AuthMiddleware())
	{
		export.GET("/clientes/excel", ctrl.ExportarClientesExcel)
		export.GET("/clientes/pdf", ctrl.ExportarClientesPDF)
		export.GET("/dispositivos/excel", ctrl.ExportarDispositivosExcel)
		export.GET("/dispositivos/pdf", ctrl.ExportarDispositivosPDF)
		export.GET("/alertas/excel", ctrl.ExportarAlertasExcel)
		export.GET("/boletas/excel", ctrl.ExportarBoletasExcel)
		export.GET("/boletas/:id/pdf", ctrl.ExportarBoletaPDF)
	}
}

func setupAuditLogRoutes(router *gin.RouterGroup, ctrl *controllers.AuditLogController) {
	audit := router.Group("/audit-logs")
	audit.Use(middleware.AuthMiddleware())
	{
		audit.GET("", ctrl.GetLogs)
		audit.GET("/user/:userId", ctrl.GetUserLogs)
		audit.GET("/resource/:resource/:resourceId", ctrl.GetResourceHistory)
		audit.GET("/statistics", ctrl.GetStatistics)
		audit.DELETE("/clean", ctrl.CleanOldLogs)
	}
}

func setupTarifaRoutes(router *gin.RouterGroup, ctrl *controllers.TarifaController) {
	tarifas := router.Group("/tarifas")
	tarifas.Use(middleware.AuthMiddleware())
	{
		tarifas.GET("", ctrl.ObtenerTodas)
		tarifas.GET("/activa", ctrl.ObtenerActiva)
		tarifas.GET("/:id", ctrl.ObtenerPorID)
		tarifas.POST("", ctrl.Crear)
		tarifas.PUT("/:id", ctrl.Actualizar)
		tarifas.DELETE("/:id", ctrl.Eliminar)
		tarifas.POST("/calcular", ctrl.CalcularConsumo)
	}
}

func setupConsumoRoutes(router *gin.RouterGroup, ctrl *controllers.ConsumoController) {
	consumo := router.Group("/consumo")
	consumo.Use(middleware.AuthMiddleware())
	{
		consumo.GET("/cliente/:clienteId/calcular", ctrl.CalcularCostoActual)
	}
}
