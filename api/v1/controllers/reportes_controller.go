package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ReportesController — controlador unificado de reportes
// Rutas:
//   GET /api/reportes/clientes?formato=excel|pdf
//   GET /api/reportes/dispositivos?formato=excel|pdf
//   GET /api/reportes/alertas?formato=excel|pdf
//   GET /api/reportes/boletas?formato=excel|pdf
//   GET /api/reportes/consumo?formato=excel|pdf&fechaInicio=YYYY-MM-DD&fechaFin=YYYY-MM-DD

type ReportesController struct {
	service *services.ReportesService
}

func NewReportesController(service *services.ReportesService) *ReportesController {
	return &ReportesController{service: service}
}

// SetupRoutes configura las rutas del controlador
func (c *ReportesController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/reportes")
	g.Use(middleware.AuthMiddleware())
	g.GET("/clientes", c.Clientes)
	g.GET("/dispositivos", c.Dispositivos)
	g.GET("/alertas", c.Alertas)
	g.GET("/boletas", c.Boletas)
	g.GET("/consumo", c.Consumo)
}

func (c *ReportesController) empresaID(gctx *gin.Context) (string, bool) {
	id := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if id == nil {
		gctx.JSON(http.StatusUnauthorized, types.ApiResponse{
			Success: false,
			Error:   "No autorizado",
		})
		return "", false
	}
	return id.(string), true
}

func (c *ReportesController) responder(gctx *gin.Context, data []byte, filename, contentType string, err error) {
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al generar reporte: " + err.Error(),
		})
		return
	}
	gctx.Header("Content-Disposition", "attachment; filename="+filename)
	gctx.Data(http.StatusOK, contentType, data)
}

func (c *ReportesController) Clientes(gctx *gin.Context) {
	empresaID, ok := c.empresaID(gctx)
	if !ok {
		return
	}
	formato := gctx.DefaultQuery("formato", "excel")
	data, filename, ct, err := c.service.ReporteClientes(gctx.Request.Context(), empresaID, formato)
	c.responder(gctx, data, filename, ct, err)
}

func (c *ReportesController) Dispositivos(gctx *gin.Context) {
	empresaID, ok := c.empresaID(gctx)
	if !ok {
		return
	}
	formato := gctx.DefaultQuery("formato", "excel")
	data, filename, ct, err := c.service.ReporteDispositivos(gctx.Request.Context(), empresaID, formato)
	c.responder(gctx, data, filename, ct, err)
}

func (c *ReportesController) Alertas(gctx *gin.Context) {
	empresaID, ok := c.empresaID(gctx)
	if !ok {
		return
	}
	formato := gctx.DefaultQuery("formato", "excel")
	data, filename, ct, err := c.service.ReporteAlertas(gctx.Request.Context(), empresaID, formato)
	c.responder(gctx, data, filename, ct, err)
}

func (c *ReportesController) Boletas(gctx *gin.Context) {
	empresaID, ok := c.empresaID(gctx)
	if !ok {
		return
	}
	formato := gctx.DefaultQuery("formato", "excel")
	data, filename, ct, err := c.service.ReporteBoletas(gctx.Request.Context(), empresaID, formato)
	c.responder(gctx, data, filename, ct, err)
}

func (c *ReportesController) Consumo(gctx *gin.Context) {
	empresaID, ok := c.empresaID(gctx)
	if !ok {
		return
	}
	formato := gctx.DefaultQuery("formato", "excel")

	fechaInicio := time.Now().AddDate(0, -1, 0)
	fechaFin := time.Now()

	if fi := gctx.Query("fechaInicio"); fi != "" {
		if t, err := time.Parse("2006-01-02", fi); err == nil {
			fechaInicio = t
		}
	}
	if ff := gctx.Query("fechaFin"); ff != "" {
		if t, err := time.Parse("2006-01-02", ff); err == nil {
			fechaFin = t
		}
	}

	data, filename, ct, err := c.service.ReporteConsumo(gctx.Request.Context(), empresaID, fechaInicio, fechaFin, formato)
	c.responder(gctx, data, filename, ct, err)
}
