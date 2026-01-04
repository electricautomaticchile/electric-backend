package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ReporteController struct {
	service *services.ReporteService
}

func NewReporteController(service *services.ReporteService) *ReporteController {
	return &ReporteController{
		service: service,
	}
}

func (ctrl *ReporteController) GenerarReporteClientes(c *gin.Context) {
	empresaID := c.Query("empresaId")
	formato := c.DefaultQuery("formato", "excel")

	if empresaID == "" {
		c.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   "empresaId es requerido",
		})
		return
	}

	buf, filename, err := ctrl.service.GenerarReporteClientes(c.Request.Context(), empresaID, formato)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al generar reporte: " + err.Error(),
		})
		return
	}

	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if formato == "csv" {
		contentType = "text/csv"
	} else if formato == "pdf" {
		contentType = "application/pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

func (ctrl *ReporteController) GenerarReporteDispositivos(c *gin.Context) {
	empresaID := c.Query("empresaId")
	formato := c.DefaultQuery("formato", "excel")

	if empresaID == "" {
		c.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   "empresaId es requerido",
		})
		return
	}

	buf, filename, err := ctrl.service.GenerarReporteDispositivos(c.Request.Context(), empresaID, formato)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al generar reporte: " + err.Error(),
		})
		return
	}

	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if formato == "csv" {
		contentType = "text/csv"
	} else if formato == "pdf" {
		contentType = "application/pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

func (ctrl *ReporteController) GenerarReporteBoletas(c *gin.Context) {
	clienteID := c.Query("clienteId")
	formato := c.DefaultQuery("formato", "excel")

	if clienteID == "" {
		c.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   "clienteId es requerido",
		})
		return
	}

	buf, filename, err := ctrl.service.GenerarReporteBoletas(c.Request.Context(), clienteID, formato)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al generar reporte: " + err.Error(),
		})
		return
	}

	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if formato == "csv" {
		contentType = "text/csv"
	} else if formato == "pdf" {
		contentType = "application/pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

func (ctrl *ReporteController) GenerarReporteConsumo(c *gin.Context) {
	empresaID := c.Query("empresaId")
	formato := c.DefaultQuery("formato", "excel")
	fechaInicioStr := c.Query("fechaInicio")
	fechaFinStr := c.Query("fechaFin")

	if empresaID == "" {
		c.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   "empresaId es requerido",
		})
		return
	}

	var fechaInicio, fechaFin time.Time
	var err error

	if fechaInicioStr != "" {
		fechaInicio, err = time.Parse("2006-01-02", fechaInicioStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ApiResponse{
				Success: false,
				Error:   "Formato de fechaInicio inválido (usar YYYY-MM-DD)",
			})
			return
		}
	} else {
		fechaInicio = time.Now().AddDate(0, -1, 0)
	}

	if fechaFinStr != "" {
		fechaFin, err = time.Parse("2006-01-02", fechaFinStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ApiResponse{
				Success: false,
				Error:   "Formato de fechaFin inválido (usar YYYY-MM-DD)",
			})
			return
		}
	} else {
		fechaFin = time.Now()
	}

	buf, filename, err := ctrl.service.GenerarReporteConsumo(c.Request.Context(), empresaID, fechaInicio, fechaFin, formato)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   "Error al generar reporte: " + err.Error(),
		})
		return
	}

	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if formato == "csv" {
		contentType = "text/csv"
	} else if formato == "pdf" {
		contentType = "application/pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, buf.Bytes())
}
