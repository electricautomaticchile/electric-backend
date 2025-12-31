package controllers

import (
	"electric-backend/domain/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EstadisticaController struct {
	estadisticaService *services.EstadisticaService
}

func NewEstadisticaController(estadisticaService *services.EstadisticaService) *EstadisticaController {
	return &EstadisticaController{
		estadisticaService: estadisticaService,
	}
}

func (ctrl *EstadisticaController) ObtenerConsumoCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	estadisticas, err := ctrl.estadisticaService.ObtenerConsumoCliente(gctx.Request.Context(), clienteID)
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
	estadisticas, err := ctrl.estadisticaService.ObtenerEstadisticasGlobales(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}
