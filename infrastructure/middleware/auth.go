package middleware

import (
	"context"
	"electric-backend/config"
	"electric-backend/types"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthSourceBearer = "bearer"
	AuthSourceCookie = "cookie"
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
		authSource := ""

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if !strings.HasPrefix(authHeader, "Bearer ") {
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
			tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			authSource = AuthSourceBearer
		} else {
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
			authSource = AuthSourceCookie
		}

		claims, err := ParseJWTClaims(tokenString)
		if err != nil {
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

		ctx := context.WithValue(c.Request.Context(), types.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, types.ContextKeyUserRole, claims.UserRole)
		ctx = context.WithValue(ctx, types.ContextKeyUserType, claims.UserType)
		ctx = context.WithValue(ctx, types.ContextKeyPowers, claims.Powers)
		ctx = context.WithValue(ctx, types.ContextKeyAuthSource, authSource)
		if claims.EmpresaID != "" {
			ctx = context.WithValue(ctx, types.ContextKeyEmpresaID, claims.EmpresaID)
		}

		c.Set("userID", claims.UserID)
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.UserRole)
		c.Set("userType", claims.UserType)
		c.Set("powers", claims.Powers)
		c.Set("authSource", authSource)
		if claims.EmpresaID != "" {
			c.Set("empresaID", claims.EmpresaID)
			c.Set("empresaId", claims.EmpresaID)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func ParseJWTClaims(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("método de firma inválido")
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("claims inválidos")
	}

	return claims, nil
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
