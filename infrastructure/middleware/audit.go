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
		if shouldSkipAudit(c.Request.URL.Path) {
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

		duration := time.Since(startTime).Milliseconds()

		var responseBody map[string]interface{}
		if blw.body.Len() > 0 && blw.body.Len() < 10000 {
			json.Unmarshal(blw.body.Bytes(), &responseBody)
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

func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/api/ws/",
		"/api/arduino/status",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

func sanitizeRequestBody(body map[string]interface{}) {
	sensitiveFields := []string{"password", "passwordTemporal", "token", "secret"}
	
	for _, field := range sensitiveFields {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
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
