package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IAController struct {
	iaService *services.IAService
}

func NewIAController(iaService *services.IAService) *IAController {
	return &IAController{
		iaService: iaService,
	}
}

func (ctrl *IAController) SetupRoutes(router *gin.RouterGroup) {
	ia := router.Group("/ia")
	{
		ia.POST("/analizar-consumo", ctrl.AnalizarConsumo)
	}
}

type BoletaAnalisis struct {
	Periodo      string                       `json:"periodo"`
	ConsumoTotal float64                      `json:"consumoTotal"`
	Monto        float64                      `json:"monto"`
	ConsumoPorHora []ConsumoPorHora           `json:"consumoPorHora,omitempty"`
}

type ConsumoPorHora struct {
	Hora     int     `json:"hora"`
	Consumo  float64 `json:"consumo"`
}

type AnalizarConsumoRequest struct {
	Boletas []BoletaAnalisis `json:"boletas"`
}

func (ctrl *IAController) AnalizarConsumo(gctx *gin.Context) {
	var req AnalizarConsumoRequest
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	if len(req.Boletas) < 3 {
		gctx.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Message: "Se requieren al menos 3 boletas para el análisis",
		})
		return
	}

	resultado, err := ctrl.iaService.AnalizarConsumo(gctx.Request.Context(), req.Boletas)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resultado,
		Message: "Análisis completado exitosamente",
	})
}
