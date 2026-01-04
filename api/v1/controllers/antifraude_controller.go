package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AntifraudeController struct {
	antifraudeService *services.AntifraudeService
}

func NewAntifraudeController(antifraudeService *services.AntifraudeService) *AntifraudeController {
	return &AntifraudeController{
		antifraudeService: antifraudeService,
	}
}

func (ctrl *AntifraudeController) DetectarAnomalias(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	anomalias, err := ctrl.antifraudeService.DetectarAnomalias(gctx.Request.Context(), empresaID.(string))
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

	estadisticas, err := ctrl.antifraudeService.ObtenerEstadisticas(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}
