package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ConsumoController struct {
	consumoService *services.ConsumoService
}

func NewConsumoController(consumoService *services.ConsumoService) *ConsumoController {
	return &ConsumoController{
		consumoService: consumoService,
	}
}

func (c *ConsumoController) CalcularCostoActual(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")
	kwhStr := gctx.Query("kwh")

	if clienteID == "" || kwhStr == "" {
		gctx.Error(types.ThrowPower("Cliente ID y kWh son requeridos"))
		return
	}

	kwh, err := strconv.ParseFloat(kwhStr, 64)
	if err != nil {
		gctx.Error(types.ThrowPower("kWh inválido"))
		return
	}

	resultado, err := c.consumoService.CalcularCostoLectura(gctx.Request.Context(), clienteID, kwh)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resultado,
	})
}
