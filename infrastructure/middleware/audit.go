package middleware

import (
	"bytes"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
	"electric-backend/types"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditMiddleware(auditService *services.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		startTime := time.Now()

		userID := ""
		userType := ""
		userEmail := ""
		empresaID := ""

		if uid := c.Request.Context().Value(types.ContextKeyUserID); uid != nil {
			userID = uid.(string)
		}
		if ut := c.Request.Context().Value(types.ContextKeyUserType); ut != nil {
			userType = ut.(string)
		}
		if eid := c.Request.Context().Value(types.ContextKeyEmpresaID); eid != nil {
			empresaID = eid.(string)
		}

		var requestBody map[string]interface{}
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) > 0 && len(bodyBytes) < 10000 {
				json.Unmarshal(bodyBytes, &requestBody)
				sanitizeRequestBody(requestBody)
			}
		}

		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		c.Next()

		userID = UserID(c)
		userType = UserType(c)
		empresaID = EmpresaID(c)

		duration := time.Since(startTime).Milliseconds()

		var responseBody map[string]interface{}
		if blw.body.Len() > 0 && blw.body.Len() < 10000 {
			json.Unmarshal(blw.body.Bytes(), &responseBody)
			// Reducir responseBody: solo guardar success, error y message
			// No guardar data completa (infla la DB innecesariamente)
			responseBody = summarizeResponseBody(responseBody)
		}

		success := c.Writer.Status() >= 200 && c.Writer.Status() < 400
		errorMessage := ""
		if !success && responseBody != nil {
			if errMap, ok := responseBody["error"].(map[string]interface{}); ok {
				if msg, ok := errMap["message"].(string); ok {
					errorMessage = msg
				}
			}
		}

		action := determineAction(c.Request.Method, c.Request.URL.Path)
		resource := determineResource(c.Request.URL.Path)
		resourceID := extractResourceID(c)

		auditLog := &models.AuditLogModel{
			UserID:       userID,
			UserType:     userType,
			UserEmail:    userEmail,
			EmpresaID:    empresaID,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			Method:       c.Request.Method,
			Endpoint:     c.Request.URL.Path,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			StatusCode:   c.Writer.Status(),
			Success:      success,
			ErrorMessage: errorMessage,
			RequestBody:  requestBody,
			ResponseBody: responseBody,
			Duration:     duration,
			Timestamp:    time.Now(),
		}

		go func() {
			auditService.Log(c.Request.Context(), auditLog)
		}()
	}
}

func shouldSkipAudit(method string, path string) bool {
	if strings.HasPrefix(path, "/api/iot/") {
		return true
	}
	if method == "POST" && path == "/api/leads" {
		return true
	}

	skipPaths := []string{
		"/health",
		"/api/arduino/status",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Filtrar escaneos de bots/atacantes que inflan la DB con basura
	scannerPatterns := []string{
		".git/", ".env", ".svn/", ".ssh/", "wp-admin", "wp-content",
		"wp-config", "phpinfo", ".DS_Store", "etc/passwd",
		"swagger-ui", "actuator/", "solr/", "graphql/console",
		"admin.php", "bolt.php", "backup.sql", ".keys/",
	}
	pathLower := strings.ToLower(path)
	for _, pattern := range scannerPatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}

	return false
}

func sanitizeRequestBody(body map[string]interface{}) {
	sensitiveFields := []string{"password", "passwordTemporal", "token", "secret", "passwordActual", "passwordNuevo"}

	for _, field := range sensitiveFields {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
		}
	}
}

// summarizeResponseBody reduce el body guardado en audit_logs.
// Solo guarda success, error y message — no el data completo que infla la DB.
func summarizeResponseBody(body map[string]interface{}) map[string]interface{} {
	if body == nil {
		return nil
	}
	summary := make(map[string]interface{})
	if v, ok := body["success"]; ok {
		summary["success"] = v
	}
	if v, ok := body["error"]; ok {
		summary["error"] = v
	}
	if v, ok := body["message"]; ok {
		summary["message"] = v
	}
	if v, ok := body["ok"]; ok {
		summary["ok"] = v
	}
	return summary
}

// MED-06: Redactar campos sensibles del response body
func sanitizeResponseBody(body map[string]interface{}) {
	if body == nil {
		return
	}
	sensitive := []string{"password", "token", "refreshToken", "rut", "correo", "telefono", "direccion"}
	for _, field := range sensitive {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
		}
	}
	if data, ok := body["data"].(map[string]interface{}); ok {
		for _, field := range sensitive {
			if _, exists := data[field]; exists {
				data[field] = "***REDACTED***"
			}
		}
	}
}

func determineAction(method string, path string) string {
	pathLower := strings.ToLower(path)

	if strings.Contains(pathLower, "/login") {
		return "login"
	}
	if strings.Contains(pathLower, "/logout") {
		return "logout"
	}
	if strings.Contains(pathLower, "/register") {
		return "register"
	}

	switch method {
	case "GET":
		if strings.Contains(pathLower, "/export") {
			return "export"
		}
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func determineResource(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) >= 2 {
		resource := parts[1]
		resource = strings.TrimSuffix(resource, "s")
		return resource
	}

	return "unknown"
}

func extractResourceID(c *gin.Context) string {
	if id := c.Param("id"); id != "" {
		return id
	}
	if id := c.Param("clienteId"); id != "" {
		return id
	}
	if id := c.Param("empresaId"); id != "" {
		return id
	}
	if id := c.Param("dispositivoId"); id != "" {
		return id
	}
	return ""
}
