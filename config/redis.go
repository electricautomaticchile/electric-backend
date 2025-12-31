package config

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// ConnectRedis conecta a Redis
func ConnectRedis(host, port, password, dbStr string) error {
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verificar la conexión
	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ No se pudo conectar a Redis: %v", err)
		log.Println("ℹ️ La aplicación continuará sin cache")
		RedisClient = nil
		return nil // No es un error crítico
	}

	log.Println("✅ Conectado a Redis")
	return nil
}

// DisconnectRedis desconecta de Redis
func DisconnectRedis() error {
	if RedisClient == nil {
		return nil
	}

	if err := RedisClient.Close(); err != nil {
		return fmt.Errorf("error desconectando de Redis: %w", err)
	}

	log.Println("✅ Desconectado de Redis")
	return nil
}
