package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportController struct {
	exportService *services.ExportService
}

func NewExportController(exportService *services.ExportService) *ExportController {
	return &ExportController{
		exportService: exportService,
	}
}

func (c *ExportController) ExportarClientesExcel(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarClientesExcel(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("clientes_%s.xlsx", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (c *ExportController) ExportarClientesPDF(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarClientesPDF(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("clientes_%s.pdf", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/pdf")
	gctx.Data(http.StatusOK, "application/pdf", data)
}

func (c *ExportController) ExportarDispositivosExcel(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarDispositivosExcel(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("dispositivos_%s.xlsx", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (c *ExportController) ExportarDispositivosPDF(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarDispositivosPDF(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("dispositivos_%s.pdf", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/pdf")
	gctx.Data(http.StatusOK, "application/pdf", data)
}

func (c *ExportController) ExportarAlertasExcel(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarAlertasExcel(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("alertas_%s.xlsx", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (c *ExportController) ExportarBoletasExcel(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	data, err := c.exportService.ExportarBoletasExcel(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("boletas_%s.xlsx", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	gctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (c *ExportController) ExportarBoletaPDF(gctx *gin.Context) {
	boletaID := gctx.Param("id")
	if boletaID == "" {
		gctx.Error(types.ThrowPower("ID de boleta requerido"))
		return
	}

	data, err := c.exportService.ExportarBoletaPDF(gctx.Request.Context(), boletaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	filename := fmt.Sprintf("boleta_%s.pdf", time.Now().Format("20060102_150405"))
	gctx.Header("Content-Description", "File Transfer")
	gctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	gctx.Header("Content-Type", "application/pdf")
	gctx.Data(http.StatusOK, "application/pdf", data)
}
