package middleware

import (
	"electric-backend/config"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura CORS
func CORSMiddleware() gin.HandlerFunc {
	// Limpiar espacios y tabs de cada origen
	rawOrigins := strings.Split(config.AppConfig.CORSOrigins, ",")
	origins := make([]string, 0, len(rawOrigins))
	for _, o := range rawOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}

	corsFn := cors.New(corsConfig)

	// Responder al preflight OPTIONS antes de cualquier otro middleware
	return func(c *gin.Context) {
		// Las rutas WebSocket no necesitan CORS — el upgrader maneja el origen
		if strings.HasPrefix(c.Request.URL.Path, "/api/ws/") {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions {
			corsFn(c)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		corsFn(c)
		c.Next()
	}
}
