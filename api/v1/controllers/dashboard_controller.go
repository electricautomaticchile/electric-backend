package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService *services.DashboardService
}

func NewDashboardController(dashboardService *services.DashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

func (ctrl *DashboardController) ObtenerEstadisticas(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	estadisticas, err := ctrl.dashboardService.ObtenerEstadisticas(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    estadisticas,
	})
}
