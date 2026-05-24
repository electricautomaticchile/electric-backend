package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func testGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c
}

func TestCanAccessResourceEmpresaScoped(t *testing.T) {
	c := testGinContext()
	c.Set("userType", "empresa")
	c.Set("empresaID", "empresa-a")

	if !CanAccessResource(c, "cliente-a", "empresa-a", false) {
		t.Fatal("empresa dueña debería acceder al recurso")
	}
	if CanAccessResource(c, "cliente-a", "empresa-b", false) {
		t.Fatal("empresa distinta no debería acceder al recurso")
	}
}

func TestCanAccessResourceClienteSelfOnly(t *testing.T) {
	c := testGinContext()
	c.Set("userType", "cliente")
	c.Set("userID", "cliente-a")

	if !CanAccessResource(c, "cliente-a", "empresa-a", true) {
		t.Fatal("cliente debería acceder a sus propios recursos cuando se permite self")
	}
	if CanAccessResource(c, "cliente-a", "empresa-a", false) {
		t.Fatal("cliente no debería acceder cuando la operación requiere empresa")
	}
	if CanAccessResource(c, "cliente-b", "empresa-a", true) {
		t.Fatal("cliente no debería acceder a recursos de otro cliente")
	}
}
