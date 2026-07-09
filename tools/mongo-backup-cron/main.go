// Backup one-shot de MongoDB a almacenamiento S3-compatible, pensado para correr
// como Render Cron Job (así el acceso a Mongo sale de IPs de Render y el allowlist
// de Atlas puede quedar restringido a Render, sin depender de AWS).
//
// Funciona con Cloudflare R2 (recomendado, off-AWS) o con AWS S3, según env:
//   MONGODB_URI            URI de Mongo.
//   MONGO_DB               base a respaldar (default electricautomaticchile).
//   S3_BUCKET              bucket destino.
//   S3_ENDPOINT            endpoint S3-compatible (R2: https://<acct>.r2.cloudflarestorage.com). Vacío = AWS S3.
//   S3_REGION              región (R2: "auto"; S3: us-east-1).
//   AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY  credenciales del storage.
//
// Sale con código 0 si todo ok; !=0 si falla (Render marca el run como fallido).
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("backup fallido: %v", err)
	}
}

func run() error {
	uri := os.Getenv("MONGODB_URI")
	bucket := os.Getenv("S3_BUCKET")
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "electricautomaticchile"
	}
	if uri == "" || bucket == "" {
		return fmt.Errorf("faltan MONGODB_URI o S3_BUCKET")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// --- Cliente S3-compatible (R2 o S3) ---
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "auto"
	}
	optFns := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if id, sec := os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"); id != "" && sec != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(id, sec, "")))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return fmt.Errorf("config AWS/R2: %w", err)
	}
	endpoint := os.Getenv("S3_ENDPOINT")
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = &endpoint
			o.UsePathStyle = true // R2 usa path-style
		}
	})

	// --- Mongo ---
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("conectando a Mongo: %w", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	db := client.Database(dbName)
	colecciones, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listando colecciones: %w", err)
	}

	ts := time.Now().UTC()
	prefix := fmt.Sprintf("backups/%s/%s", ts.Format("2006-01-02"), ts.Format("150405"))
	total := 0
	for _, nombre := range colecciones {
		cur, err := db.Collection(nombre).Find(ctx, bson.M{})
		if err != nil {
			log.Printf("omito %s: %v", nombre, err)
			continue
		}
		var docs []bson.M
		if err := cur.All(ctx, &docs); err != nil {
			log.Printf("error leyendo %s: %v", nombre, err)
			continue
		}
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		for _, d := range docs {
			b, err := bson.MarshalExtJSON(d, false, false)
			if err != nil {
				continue
			}
			_, _ = gz.Write(b)
			_, _ = gz.Write([]byte("\n"))
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("gzip %s: %w", nombre, err)
		}
		key := fmt.Sprintf("%s/%s.json.gz", prefix, nombre)
		if _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucket, Key: &key, Body: bytes.NewReader(buf.Bytes()),
		}); err != nil {
			return fmt.Errorf("subiendo %s: %w", nombre, err)
		}
		total += len(docs)
		log.Printf("backup %s: %d docs -> %s/%s", nombre, len(docs), bucket, key)
	}
	log.Printf("backup OK: %d colecciones, %d documentos, prefijo %s", len(colecciones), total, prefix)
	return nil
}
