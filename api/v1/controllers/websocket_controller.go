package controllers

import (
	"electric-backend/config"
	"electric-backend/infrastructure/websocket"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	gorillaws "github.com/gorilla/websocket"
)

// CRIT-03: Validar origen en WebSocket para prevenir CSWSH
var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Permitir conexiones sin Origin (apps nativas)
		}
		allowedOrigins := strings.Split(config.AppConfig.CORSOrigins, ",")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},
}

type wsClaims struct {
	UserID    string `json:"userId"`
	UserType  string `json:"userType"`
	EmpresaID string `json:"empresaId,omitempty"`
	jwt.RegisteredClaims
}

type WebSocketController struct {
	hub *websocket.Hub
}

func NewWebSocketController(hub *websocket.Hub) *WebSocketController {
	return &WebSocketController{hub: hub}
}

func (ctrl *WebSocketController) HandleWebSocket(gctx *gin.Context) {

	// Obtener token: cookie primero (más seguro), luego query param, luego header
	tokenStr := ""
	if c, err := gctx.Cookie("auth_token"); err == nil && c != "" {
		tokenStr = c
	}
	if tokenStr == "" {
		tokenStr = gctx.Query("token")
	}
	if tokenStr == "" {
		if h := gctx.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			tokenStr = strings.TrimPrefix(h, "Bearer ")
		}
	}

	if tokenStr == "" {
		log.Printf("❌ WS: no token provided")
		// Hacer upgrade igual pero sin autenticación — rechazar después
		gctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token requerido"})
		return
	}

	// Validar JWT directamente (sin depender del middleware)
	claims := &wsClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		gctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Hacer upgrade DESPUÉS de validar — NO escribir nada al writer después de esto
	conn, err := upgrader.Upgrade(gctx.Writer, gctx.Request, nil)
	if err != nil {
		return
	}

	client := &websocket.Client{
		Hub:       ctrl.hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    claims.UserID,
		UserType:  claims.UserType,
		EmpresaID: claims.EmpresaID,
	}

	ctrl.hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

func (ctrl *WebSocketController) GetStats(gctx *gin.Context) {
	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"connectedClients": ctrl.hub.GetConnectedClients()},
	})
}
