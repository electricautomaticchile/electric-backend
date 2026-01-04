package main

import (
	"context"
	"electric-backend/api/v1/controllers"
	"electric-backend/config"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/websocket"
	"electric-backend/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := config.ConnectDatabase(config.AppConfig.MongoURI); err != nil {
		log.Fatalf("❌ Error conectando a MongoDB: %v", err)
	}
	defer config.DisconnectDatabase()

	if err := config.ConnectRedis(
		config.AppConfig.RedisHost,
		config.AppConfig.RedisPort,
		config.AppConfig.RedisPassword,
		config.AppConfig.RedisDB,
	); err != nil {
		log.Printf("⚠️ Redis no disponible: %v", err)
	}
	defer config.DisconnectRedis()

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandler())

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
	alertaRepo := data.NewAlertaRepository()
	boletaRepo := data.NewBoletaRepository()
	ticketRepo := data.NewTicketRepository()
	configuracionRepo := data.NewConfiguracionRepository()
	estadisticaRepo := data.NewEstadisticaRepository()
	cotizacionRepo := data.NewCotizacionRepository()

	wsNotifierService := services.NewWebSocketNotifierService(wsHub)
	dashboardService := services.NewDashboardService(clienteRepo, dispositivoRepo, alertaRepo, ticketRepo)

	authService := services.NewAuthService(empresaRepo, clienteRepo, recoveryTokenRepo)
	clienteService := services.NewClienteService(clienteRepo)
	empresaService := services.NewEmpresaService(empresaRepo)
	dispositivoService := services.NewDispositivoService(dispositivoRepo, wsNotifierService)
	notificacionService := services.NewNotificacionService(notificacionRepo, wsNotifierService)
	alertaService := services.NewAlertaService(alertaRepo, wsNotifierService)
	boletaService := services.NewBoletaService(boletaRepo)
	ticketService := services.NewTicketService(ticketRepo, notificacionRepo)
	configuracionService := services.NewConfiguracionService(configuracionRepo)
	estadisticaService := services.NewEstadisticaService(estadisticaRepo)
	cotizacionService := services.NewCotizacionService(cotizacionRepo)
	alertaAutomaticaService := services.NewAlertaAutomaticaService(alertaRepo, dispositivoRepo, notificacionRepo, empresaRepo)
	reporteService := services.NewReporteService(clienteRepo, empresaRepo, dispositivoRepo, boletaRepo)
	antifraudeService := services.NewAntifraudeService(dispositivoRepo, clienteRepo, alertaRepo, notificacionRepo)

	// Inicializar facades
	authFacade := facades.NewAuthFacade(authService)
	clienteFacade := facades.NewClienteFacade(clienteService)
	empresaFacade := facades.NewEmpresaFacade(empresaService)
	dispositivoFacade := facades.NewDispositivoFacade(dispositivoService)
	cotizacionFacade := facades.NewCotizacionFacade(cotizacionService)

	// Inicializar controladores
	authController := controllers.NewAuthController(authFacade)
	clienteController := controllers.NewClienteController(clienteFacade)
	empresaController := controllers.NewEmpresaController(empresaFacade)
	dispositivoController := controllers.NewDispositivoController(dispositivoFacade)
	notificacionController := controllers.NewNotificacionController(notificacionService)
	alertaController := controllers.NewAlertaController(alertaService)
	boletaController := controllers.NewBoletaController(boletaService)
	ticketController := controllers.NewTicketController(ticketService)
	configuracionController := controllers.NewConfiguracionController(configuracionService)
	estadisticaController := controllers.NewEstadisticaController(estadisticaService)
	cotizacionController := controllers.NewCotizacionController(cotizacionFacade)
	wsController := controllers.NewWebSocketController(wsHub)
	arduinoController := controllers.NewArduinoController(arduinoBridge)
	dashboardClienteController := controllers.NewDashboardClienteController(clienteFacade, dispositivoFacade, boletaService, estadisticaService)
	alertaAutomaticaController := controllers.NewAlertaAutomaticaController(alertaAutomaticaService)
	reporteController := controllers.NewReporteController(reporteService)
	mapaController := controllers.NewMapaController(dispositivoService, clienteService)
	antifraudeController := controllers.NewAntifraudeController(antifraudeService)
	dashboardController := controllers.NewDashboardController(dashboardService)

	api := router.Group("/api")
	{
		authController.SetupRoutes(api)
		clienteController.SetupRoutes(api)
		cotizacionController.SetupRoutes(api)
		dashboardClienteController.SetupRoutes(api)

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
		setupAlertaRoutes(api, alertaController)
		setupBoletaRoutes(api, boletaController)
		setupTicketRoutes(api, ticketController)
		setupConfiguracionRoutes(api, configuracionController)
		setupEstadisticaRoutes(api, estadisticaController)
		setupAlertaAutomaticaRoutes(api, alertaAutomaticaController)
		setupReporteRoutes(api, reporteController)
		setupMapaRoutes(api, mapaController)
		setupAntifraudeRoutes(api, antifraudeController)
		setupDashboardRoutes(api, dashboardController)
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

func setupAlertaRoutes(router *gin.RouterGroup, ctrl *controllers.AlertaController) {
	alertas := router.Group("/alertas")
	alertas.Use(middleware.AuthMiddleware())
	{
		alertas.GET("", ctrl.ObtenerTodas)
		alertas.GET("/activas", ctrl.ObtenerActivas)
		alertas.GET("/empresa/:empresaId", ctrl.ObtenerPorEmpresa)
		alertas.POST("", ctrl.Crear)
		alertas.PUT("/:id/resolver", ctrl.Resolver)
		alertas.DELETE("/:id", ctrl.Eliminar)
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

func setupAlertaAutomaticaRoutes(router *gin.RouterGroup, ctrl *controllers.AlertaAutomaticaController) {
	alertasAuto := router.Group("/alertas-automaticas")
	alertasAuto.Use(middleware.AuthMiddleware())
	{
		alertasAuto.POST("/verificar/:empresaId", ctrl.VerificarManual)
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
