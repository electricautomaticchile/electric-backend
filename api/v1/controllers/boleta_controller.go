package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/export"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BoletaController struct {
	boletaService *services.BoletaService
	pdfService    *export.PDFService
}

func NewBoletaController(boletaService *services.BoletaService) *BoletaController {
	return &BoletaController{
		boletaService: boletaService,
		pdfService:    export.NewPDFService(),
	}
}

func (ctrl *BoletaController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/boletas")
	g.Use(middleware.AuthMiddleware())
	g.GET("/cliente/:clienteId", ctrl.ObtenerPorCliente)
	g.GET("/cliente/:clienteId/resumen-deuda", ctrl.ObtenerResumenDeuda)
	g.GET("/:boletaId/pdf", ctrl.GenerarPDFBoleta)
	g.POST("", ctrl.Crear)
	g.POST("/:boletaId/confirmar-pago", ctrl.ConfirmarPago)
}

func (ctrl *BoletaController) ObtenerPorCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	boletas, err := ctrl.boletaService.ObtenerPorCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    boletas,
	})
}

func (ctrl *BoletaController) ObtenerResumenDeuda(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	resumen, err := ctrl.boletaService.ObtenerResumenDeuda(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resumen,
	})
}

func (ctrl *BoletaController) Crear(gctx *gin.Context) {
	var r recipe.CrearBoletaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	boleta, err := ctrl.boletaService.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    boleta,
		"message": "Boleta creada correctamente",
	})
}

func (ctrl *BoletaController) ConfirmarPago(gctx *gin.Context) {
	boletaID := gctx.Param("boletaId")

	resultado, err := ctrl.boletaService.ConfirmarPago(gctx.Request.Context(), boletaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resultado,
		"message": resultado.Mensaje,
	})
}

func (ctrl *BoletaController) GenerarPDF(gctx *gin.Context) {
	boletaID := gctx.Param("boletaId")

	boleta, err := ctrl.boletaService.ObtenerPorID(gctx.Request.Context(), boletaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    boleta,
		"message": "PDF generado correctamente",
	})
}

func (ctrl *BoletaController) GenerarPDFBoleta(gctx *gin.Context) {
	boletaID := gctx.Param("boletaId")

	boleta, err := ctrl.boletaService.ObtenerPorID(gctx.Request.Context(), boletaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	// Obtener datos del cliente para el PDF
	clienteNombre := "Cliente"
	clienteDireccion := ""

	// Calcular tarifa por kWh
	tarifaKwh := 0.0
	if boleta.ConsumoKwh > 0 && boleta.Monto > 0 {
		tarifaKwh = boleta.Monto / boleta.ConsumoKwh
	}

	fechaEmision := boleta.FechaCreacion.Format("02/01/2006")
	fechaVencimiento := "Sin fecha"
	if boleta.FechaVencimiento != nil {
		fechaVencimiento = boleta.FechaVencimiento.Format("02/01/2006")
	}

	numeroBoleta := fmt.Sprintf("BOL-%s", boletaID[len(boletaID)-8:])

	data := map[string]interface{}{
		"numero":           numeroBoleta,
		"fechaEmision":     fechaEmision,
		"fechaVencimiento": fechaVencimiento,
		"cliente":          clienteNombre,
		"direccion":        clienteDireccion,
		"periodo":          boleta.Periodo,
		"consumo":          boleta.ConsumoKwh,
		"tarifa":           tarifaKwh,
		"monto":            boleta.Monto,
	}

	pdfBytes, err := ctrl.pdfService.ExportarBoleta(data)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error generando PDF"})
		return
	}

	filename := fmt.Sprintf("boleta-%s-%s.pdf", boleta.Periodo, time.Now().Format("20060102"))
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/pdf")
	gctx.Data(http.StatusOK, "application/pdf", pdfBytes)
}
