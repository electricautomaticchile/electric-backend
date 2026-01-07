package config

import (
	"log"
	"os"

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
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3BucketImages     string
	S3BucketDocs       string
	
	// Email
	ResendAPIKey string
	EmailFrom    string
	
	// WebSocket
	WebSocketAPIURL string
}

var AppConfig *Config

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() *Config {
	// Cargar .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "4000"),
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017/electricautomaticchile"),
		JWTSecret:   getEnv("JWT_SECRET", "secret_key_default"),
		Environment: getEnv("NODE_ENV", "development"),
		
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),
		
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),
		
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3BucketImages:     getEnv("AWS_S3_IMAGES_BUCKET_NAME", ""),
		S3BucketDocs:       getEnv("AWS_S3_BUCKET_NAME", ""),
		
		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		EmailFrom:    getEnv("EMAIL_FROM", "noreply@electricautomaticchile.com"),
		
		WebSocketAPIURL: getEnv("WEBSOCKET_API_URL", "http://localhost:5000"),
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
