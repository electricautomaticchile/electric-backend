package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
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

// SetupRoutes configura las rutas del controlador
func (ctrl *EstadisticaController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/estadisticas")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("/cliente/:clienteId", ctrl.ObtenerConsumoCliente)
	g.GET("/globales", ctrl.ObtenerEstadisticasGlobales)
	g.GET("/resumen", ctrl.ObtenerResumen)
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

// ObtenerResumen godoc
// GET /v1/estadisticas/resumen?empresaID=X
// Retorna: clientes activos, kWh totales este mes, costo promedio por cliente, alertas activas
func (ctrl *EstadisticaController) ObtenerResumen(gctx *gin.Context) {
	empresaID := gctx.Query("empresaID")

	stats, err := ctrl.dashboardService.ObtenerEstadisticas(gctx.Request.Context(), empresaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	costoPorCliente := 0.0
	if stats.ClientesActivos > 0 {
		// Estimación: tarifa BT-1A Chilquinta 284 CLP/kWh
		costoPorCliente = (stats.ConsumoTotal * 284) / float64(stats.ClientesActivos)
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"clientesActivos":     stats.ClientesActivos,
			"clientesTotales":     stats.ClientesTotales,
			"kwhTotalMes":         stats.ConsumoTotal,
			"costoPorCliente":     costoPorCliente,
			"alertasActivas":      stats.AlertasActivas,
			"dispositivosActivos": stats.DispositivosActivos,
			"dispositivosTotales": stats.DispositivosTotales,
		},
	})
}
