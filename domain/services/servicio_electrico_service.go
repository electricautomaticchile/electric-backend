package services

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/arduino"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServicioElectricoService encapsula la lógica de corte y reposición del suministro.
// Centraliza la persistencia en MongoDB y el envío de comandos al dispositivo.
type ServicioElectricoService struct {
	dispositivoRepo ports.PortDispositivo
	arduinoBridge   *arduino.SerialBridge
}

func NewServicioElectricoService(
	dispositivoRepo ports.PortDispositivo,
	arduinoBridge *arduino.SerialBridge,
) *ServicioElectricoService {
	return &ServicioElectricoService{
		dispositivoRepo: dispositivoRepo,
		arduinoBridge:   arduinoBridge,
	}
}

// CortarServicio persiste el estado "cortado" en MongoDB y envía DESACTIVAR_SERVICIO al dispositivo.
func (s *ServicioElectricoService) CortarServicio(clienteID string) {
	s.persistirEstado(clienteID, "cortado")
	s.enviarComando(clienteID, "DESACTIVAR_SERVICIO")
	log.Printf("⚡ Servicio CORTADO para cliente %s", clienteID)
}

// RestablecerServicio persiste el estado "activo" en MongoDB y envía ACTIVAR_SERVICIO al dispositivo.
func (s *ServicioElectricoService) RestablecerServicio(clienteID string) {
	s.persistirEstado(clienteID, "activo")
	s.enviarComando(clienteID, "ACTIVAR_SERVICIO")
	log.Printf("⚡ Servicio RESTABLECIDO para cliente %s", clienteID)
}

// MarcarCortePendiente persiste el estado "corte_pendiente" en MongoDB.
// El scheduler lo retomará si el servidor se reinicia antes de ejecutar el corte.
func (s *ServicioElectricoService) MarcarCortePendiente(clienteID string) {
	s.persistirEstado(clienteID, "corte_pendiente")
	log.Printf("⚡ Corte PENDIENTE marcado para cliente %s", clienteID)
}

// EjecutarCortesPendientes busca dispositivos con estado "corte_pendiente" y ejecuta el corte.
// Se llama al inicio del servidor para retomar cortes interrumpidos por reinicio.
func (s *ServicioElectricoService) EjecutarCortesPendientes(ctx context.Context) {
	if config.MongoDB == nil {
		return
	}

	collection := config.MongoDB.Collection("dispositivos")
	cursor, err := collection.Find(ctx, bson.M{"estadoServicio": "corte_pendiente"})
	if err != nil {
		log.Printf("Error buscando cortes pendientes: %v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		var clienteID string
		if oid, ok := doc["clienteId"].(primitive.ObjectID); ok {
			clienteID = oid.Hex()
		}
		if clienteID == "" {
			continue
		}

		log.Printf("⚡ Retomando corte pendiente para cliente %s", clienteID)
		s.CortarServicio(clienteID)
	}
}

func (s *ServicioElectricoService) persistirEstado(clienteID string, estado string) {
	if config.MongoDB == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(clienteID)
	filter := bson.M{"clienteId": clienteID}
	if err == nil {
		filter = bson.M{"clienteId": oid}
	}

	config.MongoDB.Collection("dispositivos").UpdateMany(ctx,
		filter,
		bson.M{"$set": bson.M{"estadoServicio": estado}},
	)
}

func (s *ServicioElectricoService) enviarComando(clienteID string, comando string) {
	if s.arduinoBridge == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispositivos, err := s.dispositivoRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return
	}

	for _, d := range dispositivos {
		if d.NumeroDispositivo != "" {
			if err := s.arduinoBridge.SendCommandToDevice(d.NumeroDispositivo, comando); err != nil {
				log.Printf("Error enviando %s a %s: %v", comando, d.NumeroDispositivo, err)
			} else {
				log.Printf("📤 %s enviado a %s", comando, d.NumeroDispositivo)
			}
			return
		}
	}

	fmt.Printf("⚠️ No se encontró dispositivo para cliente %s\n", clienteID)
}
