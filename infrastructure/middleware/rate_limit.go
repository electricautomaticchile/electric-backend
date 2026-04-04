package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"electric-backend/config"

	"github.com/gin-gonic/gin"
)

// fallbackVisitor se usa cuando Redis no está disponible
type fallbackVisitor struct {
	lastSeen time.Time
	count    int
}

var (
	fallbackVisitors = make(map[string]*fallbackVisitor)
	fallbackMu       sync.RWMutex
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			fallbackMu.Lock()
			for key, v := range fallbackVisitors {
				if time.Since(v.lastSeen) > 2*time.Minute {
					delete(fallbackVisitors, key)
				}
			}
			fallbackMu.Unlock()
		}
	}()
}

// checkRateLimitRedis intenta usar Redis para rate limiting.
// Retorna (allowed bool, err error). Si err != nil, hacer fallback a memoria.
func checkRateLimitRedis(key string, limit int, window time.Duration) (bool, error) {
	if config.RedisClient == nil {
		return false, fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	count, err := config.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		config.RedisClient.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

// checkRateLimitFallback usa el mapa en memoria como fallback
func checkRateLimitFallback(key string, limit int) bool {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	v, exists := fallbackVisitors[key]
	if !exists {
		fallbackVisitors[key] = &fallbackVisitor{lastSeen: time.Now(), count: 1}
		return true
	}

	if time.Since(v.lastSeen) > time.Minute {
		v.lastSeen = time.Now()
		v.count = 1
		return true
	}

	v.count++
	return v.count <= limit
}

func rateLimitResponse(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error": gin.H{
			"message": "Demasiadas peticiones. Por favor, intenta más tarde.",
		},
	})
	c.Abort()
}

func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("rate:%s", userID)

		allowed, err := checkRateLimitRedis(key, requestsPerMinute, time.Minute)
		if err != nil {
			// Fallback a memoria
			if !checkRateLimitFallback(key, requestsPerMinute) {
				rateLimitResponse(c)
				return
			}
			c.Next()
			return
		}

		if !allowed {
			rateLimitResponse(c)
			return
		}
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

		userID, hasUser := c.Get("userID")
		if !hasUser {
			userID = c.ClientIP()
		}

		key := fmt.Sprintf("rate:%s:%s", userID, endpoint)

		allowed, err := checkRateLimitRedis(key, limit, time.Minute)
		if err != nil {
			if !checkRateLimitFallback(key, limit) {
				rateLimitResponse(c)
				return
			}
			c.Next()
			return
		}

		if !allowed {
			rateLimitResponse(c)
			return
		}
		c.Next()
	}
}
