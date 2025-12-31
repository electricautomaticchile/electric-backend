package middleware

import (
	"electric-backend/config"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura CORS
func CORSMiddleware() gin.HandlerFunc {
	origins := strings.Split(config.AppConfig.CORSOrigins, ",")
	
	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 horas
	}

	return cors.New(corsConfig)
}
