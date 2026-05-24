package controllers

import (
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ConsumoController struct {
	consumoService    *services.ConsumoService
	dispositivoFacade *facades.DispositivoFacade
}

func NewConsumoController(consumoService *services.ConsumoService, dispositivoFacade ...*facades.DispositivoFacade) *ConsumoController {
	c := &ConsumoController{consumoService: consumoService}
	if len(dispositivoFacade) > 0 {
		c.dispositivoFacade = dispositivoFacade[0]
	}
	return c
}

// SetupRoutes configura las rutas del controlador
func (c *ConsumoController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/consumo")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("/cliente/:clienteId/calcular", c.CalcularCostoActual)
	g.GET("/cliente/:clienteId/actual", c.ObtenerConsumoActual)
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

// ObtenerConsumoActual devuelve el consumo en tiempo real del cliente
// basado en la última lectura de sus dispositivos
func (c *ConsumoController) ObtenerConsumoActual(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")
	if clienteID == "" {
		gctx.Error(types.ThrowPower("Cliente ID requerido"))
		return
	}

	if c.dispositivoFacade == nil {
		gctx.JSON(http.StatusOK, types.ApiResponse{
			Success: true,
			Data: gin.H{
				"energia":        0,
				"costo":          0,
				"potenciaActiva": 0,
			},
		})
		return
	}

	dispositivos, err := c.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	energia := 0.0
	costo := 0.0
	potencia := 0.0

	for _, d := range dispositivos {
		if d.UltimaLectura != nil {
			energia += d.UltimaLectura.Energy
			costo += d.UltimaLectura.Cost
			potencia += d.UltimaLectura.ActivePower
		}
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"energia":        energia,
			"costo":          costo,
			"potenciaActiva": potencia,
		},
	})
}
