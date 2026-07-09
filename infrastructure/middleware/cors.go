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
	isProd := config.AppConfig.Environment == "production"
	origins := make([]string, 0, len(rawOrigins))
	for _, o := range rawOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed == "" {
			continue
		}
		// En producción no permitimos orígenes de desarrollo (localhost/127.0.0.1),
		// aunque queden en CORS_ORIGINS por error.
		if isProd && (strings.Contains(trimmed, "localhost") || strings.Contains(trimmed, "127.0.0.1")) {
			continue
		}
		origins = append(origins, trimmed)
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token", "X-Client-Type"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           2 * 3600, // MED-04: Reducido a 2 horas
	}

	corsFn := cors.New(corsConfig)

	// Responder al preflight OPTIONS antes de cualquier otro middleware
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			corsFn(c)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		corsFn(c)
		c.Next()
	}
}
