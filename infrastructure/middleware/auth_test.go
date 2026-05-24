package middleware

import (
	"electric-backend/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func testJWT(t *testing.T) string {
	t.Helper()
	previous := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: "test-secret"}
	t.Cleanup(func() {
		config.AppConfig = previous
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":    "user-a",
		"userRole":  "cliente",
		"userType":  "cliente",
		"empresaId": "empresa-a",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		t.Fatalf("no se pudo firmar JWT de prueba: %v", err)
	}
	return signed
}

func TestAuthMiddlewareRejectsQueryToken(t *testing.T) {
	token := testJWT(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected?token="+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("token en query no deberia autenticar: status %d", w.Code)
	}
}

func TestAuthMiddlewareAcceptsCookieToken(t *testing.T) {
	token := testJWT(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		if c.GetString("authSource") != AuthSourceCookie {
			t.Fatalf("authSource esperado %q, obtuvo %q", AuthSourceCookie, c.GetString("authSource"))
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("cookie auth deberia autenticar: status %d", w.Code)
	}
}
