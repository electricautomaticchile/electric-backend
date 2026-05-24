package middleware

import (
	"electric-backend/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserID(c *gin.Context) string {
	if v := c.GetString("userID"); v != "" {
		return v
	}
	if v := c.GetString("userId"); v != "" {
		return v
	}
	if v := c.Request.Context().Value(types.ContextKeyUserID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func UserType(c *gin.Context) string {
	if v := c.GetString("userType"); v != "" {
		return strings.ToLower(v)
	}
	if v := c.Request.Context().Value(types.ContextKeyUserType); v != nil {
		if s, ok := v.(string); ok {
			return strings.ToLower(s)
		}
	}
	return ""
}

func UserRole(c *gin.Context) string {
	if v := c.GetString("userRole"); v != "" {
		return strings.ToLower(v)
	}
	if v := c.Request.Context().Value(types.ContextKeyUserRole); v != nil {
		if s, ok := v.(string); ok {
			return strings.ToLower(s)
		}
	}
	return ""
}

func EmpresaID(c *gin.Context) string {
	if v := c.GetString("empresaID"); v != "" {
		return v
	}
	if v := c.GetString("empresaId"); v != "" {
		return v
	}
	if v := c.Request.Context().Value(types.ContextKeyEmpresaID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func IsEmpresaContext(c *gin.Context) bool {
	userType := UserType(c)
	return userType == "empresa" || userType == "usuario_empresa"
}

func IsClienteContext(c *gin.Context) bool {
	return UserType(c) == "cliente"
}

func CanAccessEmpresa(c *gin.Context, empresaID string) bool {
	if empresaID == "" {
		return false
	}
	if isGlobalAdmin(UserRole(c)) {
		return true
	}
	return IsEmpresaContext(c) && EmpresaID(c) == empresaID
}

func CanAccessResource(c *gin.Context, resourceClienteID string, resourceEmpresaID string, allowClienteSelf bool) bool {
	if isGlobalAdmin(UserRole(c)) {
		return true
	}

	if IsClienteContext(c) {
		return allowClienteSelf && resourceClienteID != "" && resourceClienteID == UserID(c)
	}

	if IsEmpresaContext(c) {
		empresaID := EmpresaID(c)
		return empresaID != "" && resourceEmpresaID != "" && empresaID == resourceEmpresaID
	}

	return false
}

func isGlobalAdmin(role string) bool {
	switch strings.ToLower(role) {
	case "admin", "superadmin", "super_admin":
		return true
	default:
		return false
	}
}
