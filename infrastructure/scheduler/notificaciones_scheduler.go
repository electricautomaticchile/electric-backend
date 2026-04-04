package scheduler

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/services"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificacionesScheduler struct {
	smsService *services.NotificacionSMSService
	stopChan   chan bool
}

func NewNotificacionesScheduler(smsService *services.NotificacionSMSService) *NotificacionesScheduler {
	return &NotificacionesScheduler{
		smsService: smsService,
		stopChan:   make(chan bool),
	}
}

func (s *NotificacionesScheduler) Iniciar() {
	go s.ejecutarNotificacionesQuincenales()
	go s.ejecutarVerificacionBoletasImpagas()
}

func (s *NotificacionesScheduler) Detener() {
	close(s.stopChan)
}

// Mejora #11: Registrar última ejecución en Redis/MongoDB para persistencia
func (s *NotificacionesScheduler) getLastExecution(ctx context.Context, taskName string) (time.Time, error) {
	// Intentar Redis primero
	if config.RedisClient != nil {
		val, err := config.RedisClient.Get(ctx, "scheduler:"+taskName).Result()
		if err == nil {
			t, err := time.Parse(time.RFC3339, val)
			if err == nil {
				return t, nil
			}
		}
	}

	// Fallback a MongoDB
	if config.MongoDB != nil {
		collection := config.MongoDB.Collection("scheduler_state")
		var result struct {
			LastExecution time.Time `bson:"lastExecution"`
		}
		err := collection.FindOne(ctx, map[string]string{"task": taskName}).Decode(&result)
		if err == nil {
			return result.LastExecution, nil
		}
	}

	return time.Time{}, fmt.Errorf("no previous execution found")
}

func (s *NotificacionesScheduler) setLastExecution(ctx context.Context, taskName string, t time.Time) {
	// Guardar en Redis (TTL 60 días)
	if config.RedisClient != nil {
		config.RedisClient.Set(ctx, "scheduler:"+taskName, t.Format(time.RFC3339), 60*24*time.Hour)
	}

	// Guardar en MongoDB como respaldo persistente
	if config.MongoDB != nil {
		collection := config.MongoDB.Collection("scheduler_state")
		opts := options.Update().SetUpsert(true)
		collection.UpdateOne(ctx,
			map[string]string{"task": taskName},
			map[string]interface{}{
				"$set": map[string]interface{}{
					"task":          taskName,
					"lastExecution": t,
					"updatedAt":     time.Now(),
				},
			},
			opts,
		)
	}
}

func (s *NotificacionesScheduler) shouldExecute(ctx context.Context, taskName string, interval time.Duration) bool {
	lastExec, err := s.getLastExecution(ctx, taskName)
	if err != nil {
		return true // Nunca se ejecutó, ejecutar ahora
	}
	return time.Since(lastExec) >= interval
}

func (s *NotificacionesScheduler) ejecutarNotificacionesQuincenales() {
	// Verificar al inicio si hay ejecuciones pendientes por reinicio de contenedor
	ctx := context.Background()
	if s.shouldExecute(ctx, "notificaciones_quincenales", 12*time.Hour) {
		ahora := time.Now()
		dia := ahora.Day()
		if dia == 1 || dia == 15 {
			err := s.smsService.EnviarNotificacionesConsumoQuincenal(ctx)
			if err != nil {
				fmt.Printf("Error enviando notificaciones quincenales (recuperación): %v\n", err)
			} else {
				s.setLastExecution(ctx, "notificaciones_quincenales", ahora)
				fmt.Printf("Notificaciones quincenales recuperadas exitosamente el %s\n", ahora.Format("2006-01-02"))
			}
		}
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahora := time.Now()
			dia := ahora.Day()

			if dia == 1 || dia == 15 {
				ctx := context.Background()
				if !s.shouldExecute(ctx, "notificaciones_quincenales", 12*time.Hour) {
					continue
				}
				err := s.smsService.EnviarNotificacionesConsumoQuincenal(ctx)
				if err != nil {
					fmt.Printf("Error enviando notificaciones quincenales: %v\n", err)
				} else {
					s.setLastExecution(ctx, "notificaciones_quincenales", ahora)
					fmt.Printf("Notificaciones quincenales enviadas exitosamente el %s\n", ahora.Format("2006-01-02"))
				}
			}

		case <-s.stopChan:
			return
		}
	}
}

func (s *NotificacionesScheduler) ejecutarVerificacionBoletasImpagas() {
	// Verificar al inicio si hay ejecuciones pendientes
	ctx := context.Background()
	if s.shouldExecute(ctx, "verificacion_boletas", 12*time.Hour) {
		err := s.smsService.VerificarYNotificarBoletasImpagas(ctx)
		if err != nil {
			fmt.Printf("Error verificando boletas impagas (recuperación): %v\n", err)
		} else {
			s.setLastExecution(ctx, "verificacion_boletas", time.Now())
		}
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if !s.shouldExecute(ctx, "verificacion_boletas", 12*time.Hour) {
				continue
			}
			err := s.smsService.VerificarYNotificarBoletasImpagas(ctx)
			if err != nil {
				fmt.Printf("Error verificando boletas impagas: %v\n", err)
			} else {
				s.setLastExecution(ctx, "verificacion_boletas", time.Now())
			}

		case <-s.stopChan:
			return
		}
	}
}
