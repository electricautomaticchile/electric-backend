package controllers

import (
	"electric-backend/infrastructure/arduino"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ArduinoController struct {
	bridge *arduino.SerialBridge
}

func NewArduinoController(bridge *arduino.SerialBridge) *ArduinoController {
	return &ArduinoController{
		bridge: bridge,
	}
}

func (ctrl *ArduinoController) GetStatus(gctx *gin.Context) {
	connected := ctrl.bridge.IsConnected()
	devices := ctrl.bridge.GetDevices()
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"connected":      connected,
			"devicesCount":   len(devices),
			"devices":        devices,
		},
	})
}

func (ctrl *ArduinoController) ListPorts(gctx *gin.Context) {
	ports, err := ctrl.bridge.ListPorts()
	if err != nil {
		gctx.Error(types.ThrowServer("Error listando puertos"))
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"ports": ports,
		},
	})
}

func (ctrl *ArduinoController) Connect(gctx *gin.Context) {
	var req struct {
		Port string `json:"port"`
	}
	
	if err := gctx.ShouldBindJSON(&req); err != nil {
		req.Port = ""
	}
	
	if err := ctrl.bridge.Connect(req.Port); err != nil {
		gctx.Error(types.ThrowServer("Error conectando a Arduino: " + err.Error()))
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Conectado a Arduino exitosamente",
	})
}

func (ctrl *ArduinoController) Disconnect(gctx *gin.Context) {
	ctrl.bridge.Disconnect()
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Desconectado de Arduino",
	})
}

func (ctrl *ArduinoController) SendCommand(gctx *gin.Context) {
	var req struct {
		Command string `json:"command" binding:"required"`
	}
	
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.Error(types.ThrowRecipe("Comando requerido", "command"))
		return
	}
	
	if err := ctrl.bridge.SendCommand(req.Command); err != nil {
		gctx.Error(types.ThrowServer("Error enviando comando: " + err.Error()))
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Comando enviado exitosamente",
	})
}
