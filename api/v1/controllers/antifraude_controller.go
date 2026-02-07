package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AntifraudeController struct {
	monitoreoService *services.MonitoreoService
}

func NewAntifraudeController(monitoreoService *services.MonitoreoService) *AntifraudeController {
	return &AntifraudeController{
		monitoreoService: monitoreoService,
	}
}

func (ctrl *AntifraudeController) DetectarAnomalias(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	anomalias, err := ctrl.monitoreoService.DetectarAnomalias(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    anomalias,
	})
}

func (ctrl *AntifraudeController) ObtenerEstadisticas(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	estadisticas, err := ctrl.monitoreoService.ObtenerEstadisticasAntifraude(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}
