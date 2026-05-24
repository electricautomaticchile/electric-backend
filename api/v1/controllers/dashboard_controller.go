package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
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

// SetupRoutes configura las rutas del controlador
func (ctrl *DashboardController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/dashboard")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("/estadisticas", ctrl.ObtenerEstadisticas)
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
