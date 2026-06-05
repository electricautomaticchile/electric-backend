package middleware

import (
	"context"
	"electric-backend/config"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type iotTokenCacheEntry struct {
	deviceID  string
	expiresAt time.Time
}

var iotTokenCache sync.Map

// IoTAPIKeyMiddleware valida API Key para rutas IoT (CRIT-02)
// Mejora #9: Ahora valida token único por dispositivo desde MongoDB
func IoTAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-IoT-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API Key IoT no proporcionada",
			})
			c.Abort()
			return
		}

		// API key global: compatibilidad y path rapido sin roundtrip a Mongo.
		expectedKey := ""
		if config.AppConfig != nil {
			expectedKey = config.AppConfig.IoTAPIKey
		}
		if expectedKey != "" && apiKey == expectedKey {
			c.Next()
			return
		}

		if cachedDeviceID, ok := cachedIoTDevice(apiKey); ok {
			c.Set("iot_device_id", cachedDeviceID)
			c.Next()
			return
		}

		// Primero intentar validar como token único de dispositivo en MongoDB
		if config.MongoDB != nil {
			collection := config.MongoDB.Collection("dispositivos")
			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()

			var result struct {
				NumeroDispositivo string `bson:"numeroDispositivo"`
				Activo            bool   `bson:"activo"`
			}
			err := collection.FindOne(ctx, bson.M{
				"iotToken": apiKey,
				"activo":   true,
			}, options.FindOne().SetProjection(bson.M{
				"numeroDispositivo": 1,
				"activo":            1,
			})).Decode(&result)

			if err == nil {
				// Token válido de dispositivo — inyectar deviceId en contexto
				c.Set("iot_device_id", result.NumeroDispositivo)
				cacheIoTDevice(apiKey, result.NumeroDispositivo)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "API Key IoT inválida o no proporcionada",
		})
		c.Abort()
	}
}

func cachedIoTDevice(apiKey string) (string, bool) {
	value, ok := iotTokenCache.Load(apiKey)
	if !ok {
		return "", false
	}
	entry, ok := value.(iotTokenCacheEntry)
	if !ok || time.Now().After(entry.expiresAt) {
		iotTokenCache.Delete(apiKey)
		return "", false
	}
	return entry.deviceID, true
}

func cacheIoTDevice(apiKey string, deviceID string) {
	ttl := 5 * time.Minute
	if config.AppConfig != nil {
		ttl = config.AppConfig.IoTTokenCacheTTL
	}
	if ttl <= 0 || deviceID == "" {
		return
	}
	iotTokenCache.Store(apiKey, iotTokenCacheEntry{
		deviceID:  deviceID,
		expiresAt: time.Now().Add(ttl),
	})
}
