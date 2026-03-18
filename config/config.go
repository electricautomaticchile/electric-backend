package config

import (
	"log"
	"os"
	"strings"

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

// LoadConfig carga la configuración desde variables de entorno.
// En producción (ECS Fargate), el secreto completo llega como JSON en la variable ENVIRONMENT.
// En desarrollo local se usa el archivo .env.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	// ECS inyecta el secreto de Secrets Manager como JSON en la variable ENVIRONMENT.
	// Lo parseamos y seteamos cada clave como env var individual.
	if raw := os.Getenv("ENVIRONMENT"); raw != "" {
		lines := strings.Split(raw, "\n")
		for _, line := range lines {
			if !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "4000"),
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017/electricautomaticchile"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		Environment: getEnv("NODE_ENV", "development"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),

		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		S3BucketImages: getEnv("AWS_S3_IMAGES_BUCKET_NAME", ""),
		S3BucketDocs:   getEnv("AWS_S3_BUCKET_NAME", ""),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		EmailFrom:    getEnv("EMAIL_FROM", "noreply@electricautomaticchile.com"),

		InfobipAPIKey:  getEnv("INFOBIP_API_KEY", ""),
		InfobipBaseURL: getEnv("INFOBIP_BASE_URL", "https://api.infobip.com"),
		SMSFromNumber:  getEnv("SMS_FROM_NUMBER", "ElectricCL"),

		WebSocketAPIURL: getEnv("WEBSOCKET_API_URL", "http://localhost:5000"),

		IoTAPIKey: getEnv("IOT_API_KEY", ""),
	}

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
