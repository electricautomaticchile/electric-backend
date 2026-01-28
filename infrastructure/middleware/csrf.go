package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

type CSRFStore struct {
	tokens map[string]CSRFToken
	mu     sync.RWMutex
}

var csrfStore = &CSRFStore{
	tokens: make(map[string]CSRFToken),
}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *CSRFStore) Set(sessionID string, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[sessionID] = CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func (s *CSRFStore) Get(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, exists := s.tokens[sessionID]
	if !exists || time.Now().After(token.ExpiresAt) {
		return "", false
	}
	return token.Token, true
}

func (s *CSRFStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, sessionID)
}

func (s *CSRFStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for sessionID, token := range s.tokens {
		if now.After(token.ExpiresAt) {
			delete(s.tokens, sessionID)
		}
	}
}

func CSRFMiddleware() gin.HandlerFunc {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			csrfStore.Cleanup()
		}
	}()

	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		sessionID := c.GetString("session_id")
		if sessionID == "" {
			userID, exists := c.Get("userID")
			if exists {
				sessionID = userID.(string)
			}
		}

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

		serverToken, exists := csrfStore.Get(sessionID)
		if !exists || serverToken != clientToken {
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
	sessionID := c.GetString("session_id")
	if sessionID == "" {
		userID, exists := c.Get("userID")
		if exists {
			sessionID = userID.(string)
		}
	}

	if sessionID == "" {
		return ""
	}

	token, exists := csrfStore.Get(sessionID)
	if !exists {
		newToken, err := generateCSRFToken()
		if err != nil {
			return ""
		}
		csrfStore.Set(sessionID, newToken)
		return newToken
	}

	return token
}

func GenerateCSRFTokenForSession(sessionID string) (string, error) {
	token, err := generateCSRFToken()
	if err != nil {
		return "", err
	}
	csrfStore.Set(sessionID, token)
	return token, nil
}
