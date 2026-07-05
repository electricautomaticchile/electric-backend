// Lambda de backup diario de MongoDB Atlas a S3.
//
// Flujo: EventBridge (cron diario) -> esta Lambda -> S3.
//   1. Lee la MONGODB_URI desde SSM Parameter Store (SecureString).
//   2. Conecta a Mongo, recorre todas las colecciones.
//   3. Exporta cada colección a JSON extendido (una línea por documento),
//      la comprime con gzip y la sube a s3://<bucket>/backups/YYYY-MM-DD/HHMMSS/.
//
// Variables de entorno:
//   S3_BUCKET       nombre del bucket de backups.
//   SSM_PARAM_NAME  nombre del parámetro SSM con la URI (SecureString).
//   MONGO_DB        nombre de la base a respaldar (default electricautomaticchile).
//
// Módulo aislado (go.mod propio) para no afectar el build del backend.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handler(ctx context.Context) (string, error) {
	bucket := os.Getenv("S3_BUCKET")
	ssmParam := os.Getenv("SSM_PARAM_NAME")
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "electricautomaticchile"
	}
	if bucket == "" {
		return "", fmt.Errorf("falta S3_BUCKET")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("cargando config AWS: %w", err)
	}

	// 1. URI de Mongo. Orden de preferencia:
	//    a) SECRET_ID       -> AWS Secrets Manager (recomendado).
	//    b) MONGODB_URI     -> variable de entorno (cifrada en reposo).
	//    c) SSM_PARAM_NAME  -> SSM Parameter Store (SecureString).
	var uri string
	switch {
	case os.Getenv("SECRET_ID") != "":
		secretID := os.Getenv("SECRET_ID")
		sm := secretsmanager.NewFromConfig(awsCfg)
		out, err := sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &secretID})
		if err != nil {
			return "", fmt.Errorf("leyendo secreto %s: %w", secretID, err)
		}
		if out.SecretString != nil {
			uri = *out.SecretString
		}
	case os.Getenv("MONGODB_URI") != "":
		uri = os.Getenv("MONGODB_URI")
	case ssmParam != "":
		ssmClient := ssm.NewFromConfig(awsCfg)
		withDecryption := true
		param, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           &ssmParam,
			WithDecryption: &withDecryption,
		})
		if err != nil {
			return "", fmt.Errorf("leyendo SSM %s: %w", ssmParam, err)
		}
		uri = *param.Parameter.Value
	default:
		return "", fmt.Errorf("falta SECRET_ID, MONGODB_URI o SSM_PARAM_NAME")
	}
	if uri == "" {
		return "", fmt.Errorf("URI de Mongo vacía")
	}

	// 2. Conexión a Mongo.
	mctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	client, err := mongo.Connect(mctx, options.Client().ApplyURI(uri))
	if err != nil {
		return "", fmt.Errorf("conectando a Mongo: %w", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	db := client.Database(dbName)
	colecciones, err := db.ListCollectionNames(mctx, bson.M{})
	if err != nil {
		return "", fmt.Errorf("listando colecciones: %w", err)
	}

	// 3. Dump por colección a S3.
	s3Client := s3.NewFromConfig(awsCfg)
	ts := time.Now().UTC()
	prefix := fmt.Sprintf("backups/%s/%s", ts.Format("2006-01-02"), ts.Format("150405"))

	totalDocs := 0
	for _, nombre := range colecciones {
		cur, err := db.Collection(nombre).Find(mctx, bson.M{})
		if err != nil {
			log.Printf("omito %s: %v", nombre, err)
			continue
		}
		var docs []bson.M
		if err := cur.All(mctx, &docs); err != nil {
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
			return "", fmt.Errorf("gzip %s: %w", nombre, err)
		}

		key := fmt.Sprintf("%s/%s.json.gz", prefix, nombre)
		body := bytes.NewReader(buf.Bytes())
		if _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   body,
		}); err != nil {
			return "", fmt.Errorf("subiendo %s a S3: %w", nombre, err)
		}
		totalDocs += len(docs)
		log.Printf("backup %s: %d docs -> s3://%s/%s", nombre, len(docs), bucket, key)
	}

	msg := fmt.Sprintf("backup OK: %d colecciones, %d documentos, prefijo %s", len(colecciones), totalDocs, prefix)
	log.Println(msg)
	return msg, nil
}

func main() {
	lambda.Start(handler)
}
