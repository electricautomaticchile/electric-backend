package config

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Port        string
	MongoURI    string
	JWTSecret   string
	Environment string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       string

	// CORS
	CORSOrigins string

	// AWS S3
	AWSRegion      string
	S3BucketImages string
	S3BucketDocs   string

	// Email
	ResendAPIKey string
	EmailFrom    string

	// SMS Infobip
	InfobipAPIKey  string
	InfobipBaseURL string
	SMSFromNumber  string

	// WebSocket
	WebSocketAPIURL string

	// IoT
	IoTAPIKey string
}

var AppConfig *Config

// loadSecretsFromAWS obtiene secretos desde AWS Secrets Manager (solo en producción).
// Usa el IAM Role del App Runner — no requiere credenciales explícitas.
func loadSecretsFromAWS() map[string]string {
	if os.Getenv("NODE_ENV") != "production" {
		return nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("⚠️ No se pudo cargar config AWS: %v", err)
		return nil
	}

	svc := secretsmanager.NewFromConfig(cfg)
	result, err := svc.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("electric-backend/production"),
	})
	if err != nil {
		log.Printf("⚠️ No se pudo leer Secrets Manager: %v", err)
		return nil
	}

	var secrets map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		log.Printf("⚠️ Error parseando secretos: %v", err)
		return nil
	}

	log.Println("✅ Secretos cargados desde AWS Secrets Manager")
	return secrets
}

// LoadConfig carga la configuración desde Secrets Manager (producción) o variables de entorno (desarrollo)
func LoadConfig() *Config {
	// Cargar .env si existe (desarrollo local)
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	// En producción, los secretos vienen de Secrets Manager y sobreescriben las env vars
	secrets := loadSecretsFromAWS()
	getSecret := func(key, envKey, defaultValue string) string {
		if secrets != nil {
			if val, ok := secrets[key]; ok && val != "" {
				return val
			}
		}
		return getEnv(envKey, defaultValue)
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "4000"),
		MongoURI:    getSecret("MONGODB_URI", "MONGODB_URI", "mongodb://localhost:27017/electricautomaticchile"),
		JWTSecret:   getSecret("JWT_SECRET", "JWT_SECRET", ""),
		Environment: getEnv("NODE_ENV", "development"),

		RedisHost:     getSecret("REDIS_HOST", "REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getSecret("REDIS_PASSWORD", "REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),

		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		S3BucketImages: getEnv("AWS_S3_IMAGES_BUCKET_NAME", ""),
		S3BucketDocs:   getEnv("AWS_S3_BUCKET_NAME", ""),

		ResendAPIKey: getSecret("RESEND_API_KEY", "RESEND_API_KEY", ""),
		EmailFrom:    getEnv("EMAIL_FROM", "noreply@electricautomaticchile.com"),

		InfobipAPIKey:  getSecret("INFOBIP_API_KEY", "INFOBIP_API_KEY", ""),
		InfobipBaseURL: getEnv("INFOBIP_BASE_URL", "https://api.infobip.com"),
		SMSFromNumber:  getEnv("SMS_FROM_NUMBER", "ElectricCL"),

		WebSocketAPIURL: getEnv("WEBSOCKET_API_URL", "http://localhost:5000"),

		IoTAPIKey: getSecret("IOT_API_KEY", "IOT_API_KEY", ""),
	}

	// Validar que JWT_SECRET esté configurado y sea suficientemente largo
	if AppConfig.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET es requerido. Configura una clave de al menos 64 caracteres.")
	}
	if len(AppConfig.JWTSecret) < 32 {
		log.Fatal("❌ JWT_SECRET debe tener al menos 32 caracteres.")
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
