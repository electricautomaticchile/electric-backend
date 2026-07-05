package iot

import (
	"context"
	"electric-backend/config"
	"fmt"
	"time"
)

// dedupTTL define cuánto tiempo recordamos una lectura ya vista para descartar
// reenvíos. Cubre de sobra la ventana de reintentos del ESP32 tras perder
// conectividad 4G, sin crecer indefinidamente en Redis.
const dedupTTL = 15 * time.Minute

// IsDuplicateReading implementa idempotencia en la ingesta IoT usando Redis.
//
// La clave de deduplicación es (deviceID + timestamp del dispositivo). Cuando el
// ESP32 reintenta el envío de una lectura tras un corte de red, reenvía el mismo
// timestamp: la segunda vez la clave ya existe en Redis y la lectura se descarta,
// evitando duplicar consumo y métricas.
//
// Devuelve true si la lectura es un duplicado (ya vista) y debe ignorarse.
//
// Condiciones para poder deduplicar:
//   - Redis disponible (si no, no podemos deduplicar y dejamos pasar la lectura).
//   - deviceTs > 0: el dispositivo envió su propio timestamp. Si viene en 0, cada
//     reintento tomaría "now" distinto y no habría forma fiable de deduplicar, así
//     que se deja pasar.
func IsDuplicateReading(deviceID string, deviceTs int64) bool {
	if config.RedisClient == nil || deviceID == "" || deviceTs <= 0 {
		return false
	}

	key := fmt.Sprintf("iot:dedup:%s:%d", deviceID, deviceTs)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// SetNX fija la clave solo si no existía. wasSet=false => ya existía => duplicado.
	wasSet, err := config.RedisClient.SetNX(ctx, key, 1, dedupTTL).Result()
	if err != nil {
		// Ante error de Redis no bloqueamos la ingesta: preferimos aceptar la
		// lectura (posible duplicado) a perderla.
		return false
	}
	return !wasSet
}
