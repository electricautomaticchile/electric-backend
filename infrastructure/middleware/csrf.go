package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"electric-backend/config"
	"electric-backend/types"

	"github.com/gin-gonic/gin"
)

// --- Fallback en memoria (cuando Redis no está disponible) ---

type csrfTokenEntry struct {
	Token     string
	ExpiresAt time.Time
}

type csrfFallbackStore struct {
	tokens map[string]csrfTokenEntry
	mu     sync.RWMutex
}

var csrfMemStore = &csrfFallbackStore{
	tokens: make(map[string]csrfTokenEntry),
}

var csrfCleanupOnce sync.Once

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// --- Redis-backed CSRF store ---

func csrfRedisKey(sessionID string) string {
	return fmt.Sprintf("csrf:%s", sessionID)
}

func csrfSetRedis(sessionID, token string) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis not available")
	}
	ctx := context.Background()
	return config.RedisClient.Set(ctx, csrfRedisKey(sessionID), token, 24*time.Hour).Err()
}

func csrfGetRedis(sessionID string) (string, bool) {
	if config.RedisClient == nil {
		return "", false
	}
	ctx := context.Background()
	val, err := config.RedisClient.Get(ctx, csrfRedisKey(sessionID)).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// --- Fallback en memoria ---

func csrfSetMemory(sessionID, token string) {
	csrfMemStore.mu.Lock()
	defer csrfMemStore.mu.Unlock()
	csrfMemStore.tokens[sessionID] = csrfTokenEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func csrfGetMemory(sessionID string) (string, bool) {
	csrfMemStore.mu.RLock()
	defer csrfMemStore.mu.RUnlock()
	entry, exists := csrfMemStore.tokens[sessionID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return "", false
	}
	return entry.Token, true
}

func csrfCleanupMemory() {
	csrfMemStore.mu.Lock()
	defer csrfMemStore.mu.Unlock()
	now := time.Now()
	for k, v := range csrfMemStore.tokens {
		if now.After(v.ExpiresAt) {
			delete(csrfMemStore.tokens, k)
		}
	}
}

// --- Funciones unificadas (Redis con fallback a memoria) ---

func csrfSet(sessionID, token string) {
	if err := csrfSetRedis(sessionID, token); err != nil {
		csrfSetMemory(sessionID, token)
	}
}

func csrfGet(sessionID string) (string, bool) {
	if token, ok := csrfGetRedis(sessionID); ok {
		return token, true
	}
	return csrfGetMemory(sessionID)
}

// --- Middleware y helpers públicos ---

func CSRFMiddleware() gin.HandlerFunc {
	// Cleanup de memoria cada hora
	csrfCleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				csrfCleanupMemory()
			}
		}()
	})

	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		if c.GetString("authSource") != AuthSourceCookie {
			c.Next()
			return
		}

		sessionID := csrfSessionID(c)
		if sessionID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Sesión no válida",
			})
			c.Abort()
			return
		}

		clientToken := c.GetHeader("X-CSRF-Token")
		if clientToken == "" {
			clientToken = c.PostForm("csrf_token")
		}

		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Token CSRF requerido",
			})
			c.Abort()
			return
		}

		serverToken, exists := csrfGet(sessionID)
		if !exists || subtle.ConstantTimeCompare([]byte(serverToken), []byte(clientToken)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Token CSRF inválido",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetCSRFToken(c *gin.Context) string {
	sessionID := csrfSessionID(c)
	if sessionID == "" {
		return ""
	}

	token, exists := csrfGet(sessionID)
	if !exists {
		newToken, err := generateCSRFToken()
		if err != nil {
			return ""
		}
		csrfSet(sessionID, newToken)
		return newToken
	}
	return token
}

func csrfSessionID(c *gin.Context) string {
	sessionID := c.GetString("session_id")
	if sessionID == "" {
		userID, exists := c.Get("userID")
		if exists {
			sessionID = userID.(string)
		}
	}
	if sessionID == "" {
		if userID := c.GetString("userId"); userID != "" {
			sessionID = userID
		}
	}
	if sessionID == "" {
		if uid := c.Request.Context().Value(types.ContextKeyUserID); uid != nil {
			if userID, ok := uid.(string); ok {
				sessionID = userID
			}
		}
	}

	return sessionID
}

func GenerateCSRFTokenForSession(sessionID string) (string, error) {
	token, err := generateCSRFToken()
	if err != nil {
		return "", err
	}
	csrfSet(sessionID, token)
	return token, nil
}
