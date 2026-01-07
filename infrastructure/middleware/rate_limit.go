package middleware

import (
	"net/http"
	"strings"
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

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for key, v := range visitors {
				if time.Since(v.lastSeen) > 2*time.Minute {
					delete(visitors, key)
				}
			}
			mu.Unlock()
		}
	}()
}

func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			userID = c.ClientIP()
		}

		key := userID.(string)

		mu.Lock()
		v, exists := visitors[key]
		if !exists {
			visitors[key] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.lastSeen = time.Now()
			v.count = 1
			mu.Unlock()
			c.Next()
			return
		}

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

func EndpointRateLimitMiddleware(limits map[string]int) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()
		endpoint := method + ":" + path

		limit, exists := limits[endpoint]
		if !exists {
			for pattern, l := range limits {
				if strings.HasSuffix(pattern, "/*") {
					prefix := strings.TrimSuffix(pattern, "/*")
					if strings.HasPrefix(endpoint, prefix) {
						limit = l
						exists = true
						break
					}
				}
			}
		}

		if !exists {
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			userID = c.ClientIP()
		}

		key := userID.(string) + ":" + endpoint

		mu.Lock()
		v, exists := visitors[key]
		if !exists {
			visitors[key] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.lastSeen = time.Now()
			v.count = 1
			mu.Unlock()
			c.Next()
			return
		}

		v.count++
		if v.count > limit {
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
