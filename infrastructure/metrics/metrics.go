// Package metrics implementa un contador de peticiones HTTP en memoria y un
// middleware de gin para alimentarlo, sin dependencias externas. El endpoint
// /metrics (ver api/v1/server/metrics.go) lee estos contadores y los expone en
// formato de texto de exposición Prometheus.
package metrics

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	mu          sync.Mutex
	httpByCode  = map[int]uint64{}
	httpByCodeN uint64
)

// HTTPMiddleware cuenta las peticiones HTTP agrupadas por código de estado.
// Se registra una sola vez a nivel del router para tener http_requests_total.
func HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		code := c.Writer.Status()
		mu.Lock()
		httpByCode[code]++
		httpByCodeN++
		mu.Unlock()
	}
}

// HTTPRequestsSnapshot devuelve una copia del contador de peticiones por código.
func HTTPRequestsSnapshot() map[int]uint64 {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[int]uint64, len(httpByCode))
	for code, n := range httpByCode {
		out[code] = n
	}
	return out
}
