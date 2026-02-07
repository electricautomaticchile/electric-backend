package controllers

import (
	"electric-backend/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EstadisticaController struct {
	dashboardService *services.DashboardService
}

func NewEstadisticaController(dashboardService *services.DashboardService) *EstadisticaController {
	return &EstadisticaController{
		dashboardService: dashboardService,
	}
}

func (ctrl *EstadisticaController) ObtenerConsumoCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	estadisticas, err := ctrl.dashboardService.ObtenerConsumoCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}

func (ctrl *EstadisticaController) ObtenerEstadisticasGlobales(gctx *gin.Context) {
	estadisticas, err := ctrl.dashboardService.ObtenerEstadisticasGlobales(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}
