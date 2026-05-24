package controllers

import (
	"electric-backend/domain/models"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TarifaController struct {
	tarifaService *services.TarifaService
}

func NewTarifaController(tarifaService *services.TarifaService) *TarifaController {
	return &TarifaController{
		tarifaService: tarifaService,
	}
}

// SetupRoutes configura las rutas del controlador
func (c *TarifaController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/tarifas")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("", c.ObtenerTodas)
	g.GET("/activa", c.ObtenerActiva)
	g.GET("/:id", c.ObtenerPorID)
	g.POST("", c.Crear)
	g.PUT("/:id", c.Actualizar)
	g.DELETE("/:id", c.Eliminar)
	g.POST("/calcular", c.CalcularConsumo)
}

func (c *TarifaController) Crear(gctx *gin.Context) {
	var tarifa models.TarifaModel
	if err := gctx.ShouldBindJSON(&tarifa); err != nil {
		gctx.Error(types.ThrowPower("Datos inválidos: " + err.Error()))
		return
	}

	if err := c.tarifaService.CrearTarifa(gctx.Request.Context(), &tarifa); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data:    tarifa,
	})
}

func (c *TarifaController) ObtenerTodas(gctx *gin.Context) {
	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	skip := (page - 1) * limit

	tarifas, err := c.tarifaService.ObtenerTodasTarifas(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	total := len(tarifas)
	end := skip + limit
	if end > total {
		end = total
	}

	paginatedTarifas := tarifas
	if skip < total {
		paginatedTarifas = tarifas[skip:end]
	} else {
		paginatedTarifas = []*models.TarifaModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedTarifas,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (c *TarifaController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	tarifa, err := c.tarifaService.ObtenerTarifa(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    tarifa,
	})
}

func (c *TarifaController) ObtenerActiva(gctx *gin.Context) {
	comuna := gctx.Query("comuna")
	tipoTarifa := gctx.Query("tipoTarifa")

	if comuna == "" || tipoTarifa == "" {
		gctx.Error(types.ThrowPower("Comuna y tipo de tarifa son requeridos"))
		return
	}

	tarifa, err := c.tarifaService.ObtenerTarifaActiva(gctx.Request.Context(), comuna, tipoTarifa)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    tarifa,
	})
}

func (c *TarifaController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var tarifa models.TarifaModel
	if err := gctx.ShouldBindJSON(&tarifa); err != nil {
		gctx.Error(types.ThrowPower("Datos inválidos: " + err.Error()))
		return
	}

	if err := c.tarifaService.ActualizarTarifa(gctx.Request.Context(), id, &tarifa); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    tarifa,
	})
}

func (c *TarifaController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	if err := c.tarifaService.EliminarTarifa(gctx.Request.Context(), id); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    gin.H{"message": "Tarifa eliminada correctamente"},
	})
}

func (c *TarifaController) CalcularConsumo(gctx *gin.Context) {
	var req models.CalculoConsumoRequest
	if err := gctx.ShouldBindJSON(&req); err != nil {
		gctx.Error(types.ThrowPower("Datos inválidos: " + err.Error()))
		return
	}

	resultado, err := c.tarifaService.CalcularMontoConsumo(gctx.Request.Context(), req.KwhConsumidos, req.TarifaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resultado,
	})
}
