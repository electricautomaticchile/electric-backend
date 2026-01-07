package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	auditLogService *services.AuditLogService
}

func NewAuditLogController(auditLogService *services.AuditLogService) *AuditLogController {
	return &AuditLogController{
		auditLogService: auditLogService,
	}
}

func (c *AuditLogController) GetLogs(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(gctx.DefaultQuery("pageSize", "50"))
	sortBy := gctx.DefaultQuery("sortBy", "timestamp")
	sortDir := gctx.DefaultQuery("sortDir", "desc")

	params := types.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}

	customFilters := make(map[string]interface{})
	if action := gctx.Query("action"); action != "" {
		customFilters["action"] = action
	}
	if resource := gctx.Query("resource"); resource != "" {
		customFilters["resource"] = resource
	}
	if userID := gctx.Query("userId"); userID != "" {
		customFilters["userId"] = userID
	}
	if success := gctx.Query("success"); success != "" {
		customFilters["success"] = success == "true"
	}

	filters := types.FilterParams{
		Search:        gctx.Query("search"),
		DateFrom:      gctx.Query("dateFrom"),
		DateTo:        gctx.Query("dateTo"),
		CustomFilters: customFilters,
	}

	logs, err := c.auditLogService.GetLogs(gctx.Request.Context(), empresaID.(string), params, filters)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    logs,
	})
}

func (c *AuditLogController) GetUserLogs(gctx *gin.Context) {
	userID := gctx.Param("userId")
	if userID == "" {
		gctx.Error(types.ThrowPower("ID de usuario requerido"))
		return
	}

	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(gctx.DefaultQuery("pageSize", "50"))

	params := types.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   "timestamp",
		SortDir:  "desc",
	}

	logs, err := c.auditLogService.GetUserLogs(gctx.Request.Context(), userID, params)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    logs,
	})
}

func (c *AuditLogController) GetResourceHistory(gctx *gin.Context) {
	resource := gctx.Param("resource")
	resourceID := gctx.Param("resourceId")

	if resource == "" || resourceID == "" {
		gctx.Error(types.ThrowPower("Recurso y ID requeridos"))
		return
	}

	logs, err := c.auditLogService.GetResourceHistory(gctx.Request.Context(), resource, resourceID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    logs,
	})
}

func (c *AuditLogController) GetStatistics(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	startDateStr := gctx.DefaultQuery("startDate", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := gctx.DefaultQuery("endDate", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		gctx.Error(types.ThrowPower("Fecha de inicio inválida"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		gctx.Error(types.ThrowPower("Fecha de fin inválida"))
		return
	}

	stats, err := c.auditLogService.GetStatistics(gctx.Request.Context(), empresaID.(string), startDate, endDate)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    stats,
	})
}

func (c *AuditLogController) CleanOldLogs(gctx *gin.Context) {
	daysStr := gctx.DefaultQuery("days", "90")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 30 {
		gctx.Error(types.ThrowPower("Días inválidos (mínimo 30)"))
		return
	}

	count, err := c.auditLogService.CleanOldLogs(gctx.Request.Context(), days)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: map[string]interface{}{
			"deletedCount": count,
			"message":      "Logs antiguos eliminados correctamente",
		},
	})
}
