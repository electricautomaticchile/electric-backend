package main

import (
	"context"
	"electric-backend/api/v1/server"
	"electric-backend/config"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/aws"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/scheduler"
	"electric-backend/infrastructure/sms"
	"electric-backend/infrastructure/validation"
	"electric-backend/infrastructure/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// 1. Logging
	if os.Getenv("NODE_ENV") != "production" {
		log.SetOutput(&lumberjack.Logger{
			Filename: "./logs/electric-backend.log", MaxSize: 50, MaxBackups: 7, MaxAge: 30, Compress: true,
		})
	}

	// 2. Config + validators
	config.LoadConfig()
	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	registerValidators()

	// 3. Databases
	if err := config.ConnectDatabase(config.AppConfig.MongoURI); err != nil {
		log.Fatalf("❌ Error conectando a MongoDB: %v", err)
	}
	defer config.DisconnectDatabase()
	if err := config.CreateIndexes(config.MongoDB); err != nil {
		log.Printf("⚠️ Error creando índices: %v", err)
	}
	if err := config.ConnectRedis(config.AppConfig.RedisHost, config.AppConfig.RedisPort, config.AppConfig.RedisPassword, config.AppConfig.RedisDB); err != nil {
		log.Printf("⚠️ Redis no disponible: %v", err)
	}
	defer config.DisconnectRedis()

	// 4. Repositorios (build container)
	repos := data.Build()

	// 5. Servicios externos
	emailSvc := email.NewSESService()
	snsSvc, err := sms.NewSNSService()
	if err != nil {
		log.Printf("⚠️ SNS no disponible: %v", err)
	}
	s3Svc, err := aws.NewS3Service(config.AppConfig)
	if err != nil {
		log.Printf("⚠️ S3 no disponible: %v", err)
		s3Svc = nil
	}

	// 6. WebSocket + Arduino — delay aumentado para que el scheduler corra primero
	wsHub := websocket.InitializeHub()
	go wsHub.Run()
	arduinoBridge := arduino.NewSerialBridge(wsHub)
	if os.Getenv("ARDUINO_ENABLED") == "true" {
		go func() {
			// Esperar 10s para que el scheduler inicial termine antes de sincronizar el Arduino
			time.Sleep(10 * time.Second)
			port := os.Getenv("ARDUINO_PORT") // Si está vacío, auto-detecta
			if err := arduinoBridge.Connect(port); err != nil {
				log.Printf("⚠️ Arduino: %v", err)
			}
		}()
	}

	// 7. Services (build container)
	ext := &services.ExternalDeps{WSHub: wsHub, EmailSvc: emailSvc, SMSSvc: snsSvc, S3Svc: s3Svc}
	svc := services.Build(repos, ext)

	// 8. Facades (build container)
	fc := facades.Build(svc)

	// 9. Servicio eléctrico + Scheduler
	svc.BoletaService.SetDependencies(repos.DispositivoRepo, repos.NotificacionRepo, repos.TarifaRepo, snsSvc, services.NewWebSocketNotifierService(wsHub))
	servicioElectrico := services.NewServicioElectricoService(repos.DispositivoRepo, arduinoBridge)
	svc.BoletaService.SetServicioElectrico(servicioElectrico)

	// Verificación inicial SÍNCRONA — corre antes de que el Arduino se conecte
	log.Println("🔍 Verificando vencimientos de boletas al inicio...")
	if err := svc.BoletaService.VerificarVencimientos(context.Background()); err != nil {
		log.Printf("⚠️ Error en verificación inicial: %v", err)
	} else {
		log.Println("✅ Verificación inicial completada")
	}

	// Retomar cortes pendientes interrumpidos por reinicio del servidor
	servicioElectrico.EjecutarCortesPendientes(context.Background())

	notifScheduler := scheduler.NewNotificacionesScheduler(svc.NotificacionSMSService, svc.BoletaService)
	notifScheduler.Iniciar()

	// 10. Monitoreo automático
	go svc.MonitoreoService.IniciarMonitoreoAutomatico(context.Background())

	// 11. Router (server.go maneja rutas y middleware)
	router := server.SetupRouter(fc, svc, repos, wsHub, arduinoBridge)

	// 12. Start server
	port := config.AppConfig.Port
	log.Printf("✅ Servidor en puerto %s (%s)", port, config.AppConfig.Environment)
	srv := &http.Server{Addr: ":" + port, Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ %v", err)
		}
	}()

	// 13. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Apagando...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	notifScheduler.Detener()
	arduinoBridge.Disconnect()
	srv.Shutdown(shutdownCtx)
	log.Println("✅ Servidor apagado")
}

func registerValidators() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterValidation("rut", func(fl validator.FieldLevel) bool { return validation.ValidarRUT(fl.Field().String()) })
	v.RegisterValidation("password_strong", func(fl validator.FieldLevel) bool { return validation.ValidarPassword(fl.Field().String()) == nil })
	v.RegisterValidation("telefono_cl", func(fl validator.FieldLevel) bool { return validation.ValidarTelefonoChileno(fl.Field().String()) })
	v.RegisterValidation("periodo", func(fl validator.FieldLevel) bool { return validation.ValidarPeriodo(fl.Field().String()) })
	v.RegisterValidation("numero_cliente", func(fl validator.FieldLevel) bool { return validation.ValidarNumeroCliente(fl.Field().String()) })
}
