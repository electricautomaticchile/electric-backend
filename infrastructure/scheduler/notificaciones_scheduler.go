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
	smsService    *services.NotificacionSMSService
	boletaService *services.BoletaService
	stopChan      chan bool
}

func NewNotificacionesScheduler(smsService *services.NotificacionSMSService, boletaService *services.BoletaService) *NotificacionesScheduler {
	return &NotificacionesScheduler{
		smsService:    smsService,
		boletaService: boletaService,
		stopChan:      make(chan bool),
	}
}

func (s *NotificacionesScheduler) Iniciar() {
	go s.ejecutarNotificacionesQuincenales()
	go s.ejecutarVerificacionBoletasImpagas()
	go s.ejecutarVerificacionVencimientos()
	go s.ejecutarGeneracionBoletas()
	go s.ejecutarVerificacionCortesContinua() // Verifica cortes cada 5 min sin reinicio
}

func (s *NotificacionesScheduler) Detener() {
	close(s.stopChan)
}

func (s *NotificacionesScheduler) acquireTaskLock(ctx context.Context, taskName string, ttl time.Duration) bool {
	if config.RedisClient == nil {
		return true
	}
	ok, err := config.RedisClient.SetNX(ctx, "scheduler_lock:"+taskName, time.Now().Format(time.RFC3339Nano), ttl).Result()
	if err != nil {
		return false
	}
	return ok
}

func (s *NotificacionesScheduler) releaseTaskLock(ctx context.Context, taskName string) {
	if config.RedisClient == nil {
		return
	}
	config.RedisClient.Del(ctx, "scheduler_lock:"+taskName)
}

// ─── Persistencia de estado ─────────────────────────────────────────────────

func (s *NotificacionesScheduler) getLastExecution(ctx context.Context, taskName string) (time.Time, error) {
	if config.RedisClient != nil {
		val, err := config.RedisClient.Get(ctx, "scheduler:"+taskName).Result()
		if err == nil {
			t, err := time.Parse(time.RFC3339, val)
			if err == nil {
				return t, nil
			}
		}
	}

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
	if config.RedisClient != nil {
		config.RedisClient.Set(ctx, "scheduler:"+taskName, t.Format(time.RFC3339), 60*24*time.Hour)
	}

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
		return true
	}
	return time.Since(lastExec) >= interval
}

// ─── Notificaciones quincenales (día 1 y 15) ───────────────────────────────

func (s *NotificacionesScheduler) ejecutarNotificacionesQuincenales() {
	ctx := context.Background()
	if s.shouldExecute(ctx, "notificaciones_quincenales", 12*time.Hour) {
		ahora := time.Now()
		dia := ahora.Day()
		if (dia == 1 || dia == 15) && s.acquireTaskLock(ctx, "notificaciones_quincenales", 30*time.Minute) {
			err := s.smsService.EnviarNotificacionesConsumoQuincenal(ctx)
			if err != nil {
				fmt.Printf("Error enviando notificaciones quincenales (recuperación): %v\n", err)
			} else {
				s.setLastExecution(ctx, "notificaciones_quincenales", ahora)
				fmt.Printf("Notificaciones quincenales recuperadas exitosamente el %s\n", ahora.Format("2006-01-02"))
			}
			s.releaseTaskLock(ctx, "notificaciones_quincenales")
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
				if !s.acquireTaskLock(ctx, "notificaciones_quincenales", 30*time.Minute) {
					continue
				}
				err := s.smsService.EnviarNotificacionesConsumoQuincenal(ctx)
				if err != nil {
					fmt.Printf("Error enviando notificaciones quincenales: %v\n", err)
				} else {
					s.setLastExecution(ctx, "notificaciones_quincenales", ahora)
					fmt.Printf("Notificaciones quincenales enviadas el %s\n", ahora.Format("2006-01-02"))
				}
				s.releaseTaskLock(ctx, "notificaciones_quincenales")
			}
		case <-s.stopChan:
			return
		}
	}
}

// ─── Verificación de boletas impagas (SMS legacy) ───────────────────────────

func (s *NotificacionesScheduler) ejecutarVerificacionBoletasImpagas() {
	ctx := context.Background()
	if s.shouldExecute(ctx, "verificacion_boletas", 12*time.Hour) && s.acquireTaskLock(ctx, "verificacion_boletas", 30*time.Minute) {
		err := s.smsService.VerificarYNotificarBoletasImpagas(ctx)
		if err != nil {
			fmt.Printf("Error verificando boletas impagas (recuperación): %v\n", err)
		} else {
			s.setLastExecution(ctx, "verificacion_boletas", time.Now())
		}
		s.releaseTaskLock(ctx, "verificacion_boletas")
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
			if !s.acquireTaskLock(ctx, "verificacion_boletas", 30*time.Minute) {
				continue
			}
			err := s.smsService.VerificarYNotificarBoletasImpagas(ctx)
			if err != nil {
				fmt.Printf("Error verificando boletas impagas: %v\n", err)
			} else {
				s.setLastExecution(ctx, "verificacion_boletas", time.Now())
			}
			s.releaseTaskLock(ctx, "verificacion_boletas")
		case <-s.stopChan:
			return
		}
	}
}

// ─── Verificación de vencimientos (diario 08:00) ────────────────────────────

func (s *NotificacionesScheduler) ejecutarVerificacionVencimientos() {
	if s.boletaService == nil {
		return
	}

	// La verificación inicial ya la hace main.go de forma síncrona.
	// Aquí solo corremos el ticker diario.
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if !s.shouldExecute(ctx, "verificacion_vencimientos", 12*time.Hour) {
				continue
			}
			if !s.acquireTaskLock(ctx, "verificacion_vencimientos", 30*time.Minute) {
				continue
			}
			err := s.boletaService.VerificarVencimientos(ctx)
			if err != nil {
				fmt.Printf("Error verificando vencimientos: %v\n", err)
			} else {
				s.setLastExecution(ctx, "verificacion_vencimientos", time.Now())
				fmt.Printf("Vencimientos verificados el %s\n", time.Now().Format("2006-01-02"))
			}
			s.releaseTaskLock(ctx, "verificacion_vencimientos")
		case <-s.stopChan:
			return
		}
	}
}

// ─── Verificación continua de cortes (cada 5 min) ──────────────────────────
// Detecta nuevas boletas vencidas sin necesidad de reiniciar el servidor

func (s *NotificacionesScheduler) ejecutarVerificacionCortesContinua() {
	if s.boletaService == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Minute) // Verifica cortes cada 5 min
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if !s.acquireTaskLock(ctx, "verificacion_cortes_continua", 4*time.Minute) {
				continue
			}
			if err := s.boletaService.VerificarEscaladaCortes(ctx); err != nil {
				fmt.Printf("Error en verificación continua de cortes: %v\n", err)
			}
			s.releaseTaskLock(ctx, "verificacion_cortes_continua")
		case <-s.stopChan:
			return
		}
	}
}

func (s *NotificacionesScheduler) ejecutarGeneracionBoletas() {
	if s.boletaService == nil {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahora := time.Now()
			// Verificar si es el último día del mes
			manana := ahora.AddDate(0, 0, 1)
			if manana.Day() == 1 {
				ctx := context.Background()
				if !s.shouldExecute(ctx, "generacion_boletas", 20*time.Hour) {
					continue
				}
				if !s.acquireTaskLock(ctx, "generacion_boletas", 60*time.Minute) {
					continue
				}
				fmt.Printf("Generando boletas mensuales — %s\n", ahora.Format("2006-01-02"))
				err := s.boletaService.GenerarBoletasMensuales(ctx)
				if err != nil {
					fmt.Printf("Error generando boletas: %v\n", err)
				} else {
					s.setLastExecution(ctx, "generacion_boletas", ahora)
					fmt.Printf("Boletas generadas exitosamente el %s\n", ahora.Format("2006-01-02"))
				}
				s.releaseTaskLock(ctx, "generacion_boletas")
			}
		case <-s.stopChan:
			return
		}
	}
}
