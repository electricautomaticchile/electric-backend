package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI es requerido")
	}
	dbName := getenv("MONGODB_DATABASE", "electricautomaticchile_loadtest")
	count := atoi(getenv("DEVICE_COUNT", "10000"))
	prefix := getenv("DEVICE_PREFIX", "MED")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoURI).
		SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}).
		SetServerSelectionTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database(dbName).Collection("dispositivos")
	now := time.Now().UTC()
	empresaID := primitive.NewObjectID()
	clienteID := primitive.NewObjectID()

	models := make([]mongo.WriteModel, 0, count)
	for i := 0; i < count; i++ {
		numero := fmt.Sprintf("%s-%06d", prefix, i)
		doc := bson.M{
			"numeroDispositivo":  numero,
			"nombre":             "Medidor carga " + numero,
			"tipo":               "medidor",
			"clienteId":          clienteID,
			"empresaId":          empresaID,
			"estado":             "activo",
			"estadoServicio":     "activo",
			"activo":             true,
			"fechaCreacion":      now,
			"fechaActualizacion": now,
			"iotToken":           fmt.Sprintf("loadtest-token-%06d", i),
		}
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"numeroDispositivo": numero}).
			SetUpdate(bson.M{"$set": doc}).
			SetUpsert(true),
		)
	}

	result, err := collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("seed completo db=%s dispositivos=%d upserted=%d modified=%d", dbName, count, result.UpsertedCount, result.ModifiedCount)
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func atoi(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 1
	}
	return n
}
