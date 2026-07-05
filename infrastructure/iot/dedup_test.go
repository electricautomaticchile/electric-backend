package iot

import (
	"context"
	"os"
	"testing"
	"time"

	"electric-backend/config"

	"github.com/redis/go-redis/v9"
)

// TestIsDuplicateReading_SinRedis: sin Redis no se puede deduplicar, así que
// nunca marca duplicado (preferimos aceptar la lectura a perderla).
func TestIsDuplicateReading_SinRedis(t *testing.T) {
	config.RedisClient = nil
	if IsDuplicateReading("dev-1", time.Now().Unix()) {
		t.Fatal("sin Redis no debería marcar duplicado")
	}
}

// TestIsDuplicateReading_SinTimestamp: sin timestamp del dispositivo no se puede
// deduplicar de forma fiable, así que se deja pasar.
func TestIsDuplicateReading_SinTimestamp(t *testing.T) {
	config.RedisClient = nil
	if IsDuplicateReading("dev-1", 0) {
		t.Fatal("con timestamp 0 no debería marcar duplicado")
	}
}

// TestIsDuplicateReading_Reintento: simula el reenvío del ESP32. Con Redis, la
// primera lectura pasa y el reenvío (mismo deviceId+timestamp) se detecta como
// duplicado. Requiere Redis en TEST_REDIS_ADDR (por defecto localhost:6390);
// hace skip si no está disponible (CI-safe).
func TestIsDuplicateReading_Reintento(t *testing.T) {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6390"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis no disponible en %s: %v", addr, err)
	}

	config.RedisClient = client
	t.Cleanup(func() {
		config.RedisClient = nil
		_ = client.Close()
	})

	deviceID := "dev-smoke-test"
	ts := time.Now().UnixNano() // único por corrida para no chocar con datos previos

	// Primera recepción: no es duplicado.
	if IsDuplicateReading(deviceID, ts) {
		t.Fatal("la primera lectura no debería ser duplicado")
	}

	// Reenvío del mismo (deviceId + timestamp): sí es duplicado.
	if !IsDuplicateReading(deviceID, ts) {
		t.Fatal("el reenvío con el mismo timestamp debería detectarse como duplicado")
	}

	// Otra lectura con timestamp distinto: no es duplicado.
	if IsDuplicateReading(deviceID, ts+1) {
		t.Fatal("una lectura con distinto timestamp no debería ser duplicado")
	}
}
