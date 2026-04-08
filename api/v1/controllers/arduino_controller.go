package controllers

import (
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/middleware"
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

// SetupRoutes configura las rutas del controlador
func (ctrl *ArduinoController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/arduino")
	g.Use(middleware.AuthMiddleware())
	g.GET("/status", ctrl.GetStatus)
	g.GET("/ports", ctrl.ListPorts)
	g.POST("/connect", ctrl.Connect)
	g.POST("/disconnect", ctrl.Disconnect)
	g.POST("/command", ctrl.SendCommand)
}

func (ctrl *ArduinoController) GetStatus(gctx *gin.Context) {
	connected := ctrl.bridge.IsConnected()
	devices := ctrl.bridge.GetDevices()

	transformedDevices := make([]gin.H, 0, len(devices))
	for _, device := range devices {
		deviceData := gin.H{
			"ID":        device.ID,
			"ClienteID": device.ClienteID,
			"EmpresaID": device.EmpresaID,
		}

		if device.LastReading != nil {
			deviceData["LastReading"] = gin.H{
				"idDispositivo":  device.LastReading.DeviceID,
				"voltaje":        device.LastReading.Voltage,
				"corriente":      device.LastReading.Current,
				"potenciaActiva": device.LastReading.Power,
				"energia":        device.LastReading.Energy,
				"costo":          device.LastReading.Cost,
				"servicioActivo": device.LastReading.ServicioActivo,
				"uptime":         device.LastReading.Uptime,
				"marcaTiempo":    device.LastReading.Timestamp,
			}
		}

		transformedDevices = append(transformedDevices, deviceData)
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"connected":    connected,
			"devicesCount": len(devices),
			"devices":      transformedDevices,
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
		Command  string `json:"command" binding:"required"`
		DeviceID string `json:"deviceId"`
	}

	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.Error(types.ThrowRecipe("Comando requerido", "command"))
		return
	}

	if err := ctrl.bridge.SendCommandToDevice(req.DeviceID, req.Command); err != nil {
		gctx.Error(types.ThrowServer("Error enviando comando: " + err.Error()))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Comando enviado exitosamente",
	})
}
