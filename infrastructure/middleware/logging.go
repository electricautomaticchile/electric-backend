package middleware

import (
	"electric-backend/infrastructure/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger registra cada petición HTTP con zerolog de forma estructurada
// (método, ruta, status, latencia y contexto de usuario si está disponible).
// Reemplaza al logger por defecto de Gin para tener logs JSON en producción.
func RequestLogger(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := skip[path]; ok {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()
		evt := logger.Info()
		if status >= 500 {
			evt = logger.Error()
		} else if status >= 400 {
			evt = logger.Warn()
		}

		evt = evt.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP())

		// Enriquecer con contexto de usuario si el auth middleware lo dejó.
		if v, ok := c.Get("userId"); ok {
			evt = evt.Interface("userId", v)
		}
		if v, ok := c.Get("empresaId"); ok {
			evt = evt.Interface("empresaId", v)
		}

		evt.Msg("http_request")
	}
}
