package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlertaAutomaticaController struct {
	service *services.AlertaAutomaticaService
}

func NewAlertaAutomaticaController(service *services.AlertaAutomaticaService) *AlertaAutomaticaController {
	return &AlertaAutomaticaController{
		service: service,
	}
}

func (ctrl *AlertaAutomaticaController) VerificarManual(c *gin.Context) {
	empresaID := c.Param("empresaId")
	
	if err := ctrl.service.VerificarManual(c.Request.Context(), empresaID); err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al verificar alertas: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"mensaje": "Verificación de alertas completada",
		},
	})
}
