package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.RWMutex
)

// RateLimitMiddleware limita las peticiones por IP
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// Limpiar visitantes antiguos cada minuto
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			mu.Unlock()
			c.Next()
			return
		}

		// Si ha pasado más de un minuto, resetear contador
		if time.Since(v.lastSeen) > time.Minute {
			v.lastSeen = time.Now()
			v.count = 1
			mu.Unlock()
			c.Next()
			return
		}

		// Incrementar contador
		v.count++
		if v.count > requestsPerMinute {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"message": "Demasiadas peticiones. Por favor, intenta más tarde.",
				},
			})
			c.Abort()
			return
		}

		mu.Unlock()
		c.Next()
	}
}
