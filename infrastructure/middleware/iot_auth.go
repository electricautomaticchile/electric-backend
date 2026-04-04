package middleware

import (
	"context"
	"electric-backend/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

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
			}).Decode(&result)

			if err == nil {
				// Token válido de dispositivo — inyectar deviceId en contexto
				c.Set("iot_device_id", result.NumeroDispositivo)
				c.Next()
				return
			}
		}

		// Fallback: validar contra API Key global (compatibilidad)
		expectedKey := config.AppConfig.IoTAPIKey
		if expectedKey == "" || apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API Key IoT inválida o no proporcionada",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
