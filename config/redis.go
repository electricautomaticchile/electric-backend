package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// ConnectRedis conecta a un servicio compatible con Redis.
func ConnectRedis(host, port, password, dbStr string) error {
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	}

	// Usar REDIS_TLS=true para proveedores que exigen rediss/TLS.
	if os.Getenv("REDIS_TLS") == "true" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		}
	}

	RedisClient = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️ No se pudo conectar a Redis (%s:%s): %v", host, port, err)
		RedisClient = nil
		return err
	}

	log.Printf("✅ Conectado a Redis en %s:%s", host, port)
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
