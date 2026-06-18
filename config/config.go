package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	MongoMaxPool  uint64
	MongoMinPool  uint64
	JWTSecret     string
	Environment   string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       string

	// CORS
	CORSOrigins string

	// Auth cookies
	AuthCookieDomain string

	// AWS S3
	AWSRegion      string
	S3BucketImages string
	S3BucketDocs   string
	S3PublicURL    string

	// Email
	EmailFrom string

	// SMS Infobip
	InfobipAPIKey  string
	InfobipBaseURL string
	SMSFromNumber  string

	// WebSocket
	WebSocketAPIURL string

	// IoT
	IoTAPIKey              string
	IoTTokenCacheTTL       time.Duration
	IoTIngestAsync         bool
	IoTIngestQueueSize     int
	IoTIngestBatchSize     int
	IoTIngestWorkers       int
	IoTIngestFlushInterval time.Duration
	IoTIngestWriteTimeout  time.Duration
	IoTIngestMaxRetries    int

	// Landing leads ingestion
	LeadIngestAsync         bool
	LeadIngestQueueSize     int
	LeadIngestBatchSize     int
	LeadIngestWorkers       int
	LeadIngestFlushInterval time.Duration
	LeadIngestWriteTimeout  time.Duration
	LeadIngestMaxRetries    int
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
		Port:          getEnv("PORT", "4000"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017/electricautomaticchile"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "electricautomaticchile"),
		MongoMaxPool:  uint64(getEnvInt("MONGODB_MAX_POOL_SIZE", 800)),
		MongoMinPool:  uint64(getEnvInt("MONGODB_MIN_POOL_SIZE", 20)),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		Environment:   getEnv("NODE_ENV", "development"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnv("REDIS_DB", "0"),

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),

		AuthCookieDomain: getEnv("AUTH_COOKIE_DOMAIN", ""),

		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		S3BucketImages: getEnv("AWS_S3_IMAGES_BUCKET_NAME", ""),
		S3BucketDocs:   getEnv("AWS_S3_BUCKET_NAME", ""),
		S3PublicURL:    getEnv("S3_PUBLIC_URL", ""),

		EmailFrom: getEnv("EMAIL_FROM", "noreply@electricautomaticchile.com"),

		InfobipAPIKey:  getEnv("INFOBIP_API_KEY", ""),
		InfobipBaseURL: getEnv("INFOBIP_BASE_URL", "https://api.infobip.com"),
		SMSFromNumber:  getEnv("SMS_FROM_NUMBER", "ElectricCL"),

		WebSocketAPIURL: getEnv("WEBSOCKET_API_URL", "http://localhost:5000"),

		IoTAPIKey:              getEnv("IOT_API_KEY", ""),
		IoTTokenCacheTTL:       getEnvDurationMS("IOT_TOKEN_CACHE_TTL_MS", 5*time.Minute),
		IoTIngestAsync:         getEnvBool("IOT_INGEST_ASYNC", true),
		IoTIngestQueueSize:     getEnvInt("IOT_INGEST_QUEUE_SIZE", 100000),
		IoTIngestBatchSize:     getEnvInt("IOT_INGEST_BATCH_SIZE", 1000),
		IoTIngestWorkers:       getEnvInt("IOT_INGEST_WORKERS", 4),
		IoTIngestFlushInterval: getEnvDurationMS("IOT_INGEST_FLUSH_INTERVAL_MS", 100*time.Millisecond),
		IoTIngestWriteTimeout:  getEnvDurationMS("IOT_INGEST_WRITE_TIMEOUT_MS", 15*time.Second),
		IoTIngestMaxRetries:    getEnvInt("IOT_INGEST_MAX_RETRIES", 3),

		LeadIngestAsync:         getEnvBool("LEAD_INGEST_ASYNC", true),
		LeadIngestQueueSize:     getEnvInt("LEAD_INGEST_QUEUE_SIZE", 50000),
		LeadIngestBatchSize:     getEnvInt("LEAD_INGEST_BATCH_SIZE", 500),
		LeadIngestWorkers:       getEnvInt("LEAD_INGEST_WORKERS", 4),
		LeadIngestFlushInterval: getEnvDurationMS("LEAD_INGEST_FLUSH_INTERVAL_MS", 100*time.Millisecond),
		LeadIngestWriteTimeout:  getEnvDurationMS("LEAD_INGEST_WRITE_TIMEOUT_MS", 15*time.Second),
		LeadIngestMaxRetries:    getEnvInt("LEAD_INGEST_MAX_RETRIES", 3),
	}
	applyRedisURL(AppConfig)

	if AppConfig.JWTSecret == "" {
		log.Fatal("❌ JWT_SECRET es requerido. Configura una clave de al menos 64 caracteres.")
	}
	if len(AppConfig.JWTSecret) < 32 {
		log.Fatal("❌ JWT_SECRET debe tener al menos 32 caracteres.")
	}
	if AppConfig.S3PublicURL == "" {
		log.Fatal("❌ S3_PUBLIC_URL es requerido. Configura la URL de CloudFront.")
	}

	return AppConfig
}

func applyRedisURL(cfg *Config) {
	raw := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if raw == "" {
		return
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		log.Printf("⚠️ REDIS_URL inválida, usando REDIS_HOST/REDIS_PORT: %v", err)
		return
	}

	if host := parsed.Hostname(); host != "" {
		cfg.RedisHost = host
	}
	if port := parsed.Port(); port != "" {
		cfg.RedisPort = port
	}
	if password, ok := parsed.User.Password(); ok {
		cfg.RedisPassword = password
	}
	if db := strings.Trim(parsed.Path, "/"); db != "" {
		cfg.RedisDB = db
	}
	if parsed.Scheme == "rediss" && os.Getenv("REDIS_TLS") == "" {
		os.Setenv("REDIS_TLS", "true")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return defaultValue
	}
	return parsed
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvDurationMS(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return defaultValue
	}
	return time.Duration(parsed) * time.Millisecond
}
