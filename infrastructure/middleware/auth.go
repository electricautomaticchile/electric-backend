package middleware

import (
	"context"
	"electric-backend/config"
	"electric-backend/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims representa los claims del JWT
type JWTClaims struct {
	UserID    string   `json:"userId"`
	UserRole  string   `json:"userRole"`
	UserType  string   `json:"userType"`
	EmpresaID string   `json:"empresaId,omitempty"`
	Powers    []string `json:"powers,omitempty"`
	jwt.RegisteredClaims
}

// AuthMiddleware verifica el token JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Intentar obtener token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"type":    types.ErrorTypeAuth,
						"message": "Formato de token inválido",
					},
				})
				c.Abort()
				return
			}
		} else if q := c.Query("token"); q != "" {
			// Query param para WebSocket (no puede enviar headers)
			tokenString = q
		} else {
			// Si no hay header, intentar obtener token de la cookie
			cookie, err := c.Cookie("auth_token")
			if err != nil || cookie == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"type":    types.ErrorTypeAuth,
						"message": "Token no proporcionado",
					},
				})
				c.Abort()
				return
			}
			tokenString = cookie
		}

		// Verificar el token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"type":    types.ErrorTypeAuth,
					"message": "Token inválido o expirado",
				},
			})
			c.Abort()
			return
		}

		// Extraer claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"type":    types.ErrorTypeAuth,
					"message": "Claims inválidos",
				},
			})
			c.Abort()
			return
		}

		// Guardar información del usuario en el contexto
		ctx := context.WithValue(c.Request.Context(), types.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, types.ContextKeyUserRole, claims.UserRole)
		ctx = context.WithValue(ctx, types.ContextKeyUserType, claims.UserType)
		ctx = context.WithValue(ctx, types.ContextKeyPowers, claims.Powers)
		if claims.EmpresaID != "" {
			ctx = context.WithValue(ctx, types.ContextKeyEmpresaID, claims.EmpresaID)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RequireRole verifica que el usuario tenga un rol específico
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.Request.Context().Value(types.ContextKeyUserRole)
		if userRole == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"type":    types.ErrorTypePower,
					"message": "No tienes permisos para acceder a este recurso",
				},
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		hasRole := false
		for _, role := range roles {
			if roleStr == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"type":    types.ErrorTypePower,
					"message": "No tienes permisos para acceder a este recurso",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
