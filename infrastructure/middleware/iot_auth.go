package middleware

import (
	"electric-backend/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IoTAPIKeyMiddleware valida API Key para rutas IoT (CRIT-02)
func IoTAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-IoT-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

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
