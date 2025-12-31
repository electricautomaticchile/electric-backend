package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BoletaController struct {
	boletaService *services.BoletaService
}

func NewBoletaController(boletaService *services.BoletaService) *BoletaController {
	return &BoletaController{
		boletaService: boletaService,
	}
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
