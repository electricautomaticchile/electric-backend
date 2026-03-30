package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Database

// ConnectDatabase conecta a MongoDB
func ConnectDatabase(mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// HIGH-06: Configurar TLS explícito y timeouts adecuados
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(300).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error conectando a MongoDB: %w", err)
	}

	// Verificar la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("error haciendo ping a MongoDB: %w", err)
	}

	// Obtener el nombre de la base de datos desde la URI
	dbName := "electricautomaticchile"
	MongoDB = client.Database(dbName)

	log.Printf("✅ Conectado a MongoDB: %s", dbName)
	return nil
}

// DisconnectDatabase desconecta de MongoDB
func DisconnectDatabase() error {
	if MongoDB == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := MongoDB.Client().Disconnect(ctx); err != nil {
		return fmt.Errorf("error desconectando de MongoDB: %w", err)
	}

	log.Println("✅ Desconectado de MongoDB")
	return nil
}
