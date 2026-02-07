package scheduler

import (
	"context"
	"electric-backend/domain/services"
	"fmt"
	"time"
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

func (s *NotificacionesScheduler) ejecutarNotificacionesQuincenales() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahora := time.Now()
			dia := ahora.Day()

			if dia == 1 || dia == 15 {
				ctx := context.Background()
				err := s.smsService.EnviarNotificacionesConsumoQuincenal(ctx)
				if err != nil {
					fmt.Printf("Error enviando notificaciones quincenales: %v\n", err)
				} else {
					fmt.Printf("Notificaciones quincenales enviadas exitosamente el %s\n", ahora.Format("2006-01-02"))
				}
			}

		case <-s.stopChan:
			return
		}
	}
}

func (s *NotificacionesScheduler) ejecutarVerificacionBoletasImpagas() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			err := s.smsService.VerificarYNotificarBoletasImpagas(ctx)
			if err != nil {
				fmt.Printf("Error verificando boletas impagas: %v\n", err)
			}

		case <-s.stopChan:
			return
		}
	}
}
