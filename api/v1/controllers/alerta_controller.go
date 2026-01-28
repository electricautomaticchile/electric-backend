package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlertaController struct {
	alertaService *services.AlertaService
}

func NewAlertaController(alertaService *services.AlertaService) *AlertaController {
	return &AlertaController{
		alertaService: alertaService,
	}
}

func (ctrl *AlertaController) ObtenerTodas(gctx *gin.Context) {
	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	
	skip := (page - 1) * limit
	
	alertas, err := ctrl.alertaService.ObtenerTodas(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}
	
	total := len(alertas)
	end := skip + limit
	if end > total {
		end = total
	}
	
	paginatedAlertas := alertas
	if skip < total {
		paginatedAlertas = alertas[skip:end]
	} else {
		paginatedAlertas = []*models.AlertaModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedAlertas,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (ctrl *AlertaController) ObtenerActivas(gctx *gin.Context) {
	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	
	skip := (page - 1) * limit
	
	alertas, err := ctrl.alertaService.ObtenerActivas(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}
	
	total := len(alertas)
	end := skip + limit
	if end > total {
		end = total
	}
	
	paginatedAlertas := alertas
	if skip < total {
		paginatedAlertas = alertas[skip:end]
	} else {
		paginatedAlertas = []*models.AlertaModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedAlertas,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (ctrl *AlertaController) ObtenerPorEmpresa(gctx *gin.Context) {
	empresaID := gctx.Param("empresaId")

	alertas, err := ctrl.alertaService.ObtenerPorEmpresa(gctx.Request.Context(), empresaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alertas,
	})
}

func (ctrl *AlertaController) Crear(gctx *gin.Context) {
	var r recipe.CrearAlertaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	alerta, err := ctrl.alertaService.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    alerta,
		"message": "Alerta creada correctamente",
	})
}

func (ctrl *AlertaController) Resolver(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ResolverAlertaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.alertaService.Resolver(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Alerta resuelta correctamente",
})
}

func (ctrl *AlertaController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.alertaService.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Alerta eliminada correctamente",
})
}
