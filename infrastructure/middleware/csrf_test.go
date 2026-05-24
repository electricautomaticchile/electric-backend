package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func csrfTestRouter(authSource string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("authSource", authSource)
		c.Set("userID", "user-a")
		c.Next()
	})
	router.Use(CSRFMiddleware())
	router.POST("/mutate", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func TestCSRFSkipeBearerRequests(t *testing.T) {
	router := csrfTestRouter(AuthSourceBearer)
	req := httptest.NewRequest(http.MethodPost, "/mutate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Bearer no debería requerir CSRF: status %d", w.Code)
	}
}

func TestCSRFRequiresTokenForCookieAuth(t *testing.T) {
	router := csrfTestRouter(AuthSourceCookie)
	req := httptest.NewRequest(http.MethodPost, "/mutate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("cookie auth sin CSRF debería fallar: status %d", w.Code)
	}
}

func TestCSRFAcceptsValidCookieToken(t *testing.T) {
	token, err := GenerateCSRFTokenForSession("user-a")
	if err != nil {
		t.Fatalf("no se pudo generar CSRF: %v", err)
	}

	router := csrfTestRouter(AuthSourceCookie)
	req := httptest.NewRequest(http.MethodPost, "/mutate", nil)
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("CSRF válido debería pasar: status %d", w.Code)
	}
}
