package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificacionSMSController struct {
	service *services.NotificacionSMSService
}

func NewNotificacionSMSController(service *services.NotificacionSMSService) *NotificacionSMSController {
	return &NotificacionSMSController{
		service: service,
	}
}

func (ctrl *NotificacionSMSController) EnviarNotificacionManual(c *gin.Context) {
	var req struct {
		ClienteID string `json:"clienteId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   "Datos inválidos",
		})
		return
	}

	err := ctrl.service.EnviarNotificacionesConsumoQuincenal(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error enviando notificación SMS",
		})
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Notificación SMS enviada exitosamente",
	})
}

func (ctrl *NotificacionSMSController) VerificarBoletasImpagas(c *gin.Context) {
	err := ctrl.service.VerificarYNotificarBoletasImpagas(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error verificando boletas impagas",
		})
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Verificación de boletas completada",
	})
}
