package controllers

import (
	"electric-backend/infrastructure/websocket"
	"electric-backend/types"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketController struct {
	hub *websocket.Hub
}

func NewWebSocketController(hub *websocket.Hub) *WebSocketController {
	return &WebSocketController{
		hub: hub,
	}
}

func (ctrl *WebSocketController) HandleWebSocket(gctx *gin.Context) {
	token := gctx.Query("token")
	if token == "" {
		tokenCookie, err := gctx.Cookie("auth_token")
		if err == nil {
			token = tokenCookie
		}
	}

	if token == "" {
		gctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token requerido"})
		return
	}

	userID := gctx.Request.Context().Value(types.ContextKeyUserID)
	userType := gctx.Request.Context().Value(types.ContextKeyUserType)
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)

	if userID == nil || userType == nil {
		gctx.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	conn, err := upgrader.Upgrade(gctx.Writer, gctx.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &websocket.Client{
		Hub:       ctrl.hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    userID.(string),
		UserType:  userType.(string),
	}

	if empresaID != nil {
		client.EmpresaID = empresaID.(string)
	}

	ctrl.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (ctrl *WebSocketController) GetStats(gctx *gin.Context) {
	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"connectedClients": ctrl.hub.GetConnectedClients(),
		},
	})
}
