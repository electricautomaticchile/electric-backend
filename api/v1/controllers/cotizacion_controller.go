package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CotizacionController struct {
	cotizacionFacade *facades.CotizacionFacade
}

func NewCotizacionController(cotizacionFacade *facades.CotizacionFacade) *CotizacionController {
	return &CotizacionController{
		cotizacionFacade: cotizacionFacade,
	}
}

// SetupRoutes configura las rutas del controlador
func (ctrl *CotizacionController) SetupRoutes(router *gin.RouterGroup) {
	cotizaciones := router.Group("/cotizaciones")
	{
		// Ruta pública para crear cotización (desde formulario web)
		cotizaciones.POST("", ctrl.Crear)

		// Rutas protegidas
		cotizaciones.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
		{
			cotizaciones.GET("", ctrl.ObtenerTodas)
			cotizaciones.GET("/:id", ctrl.ObtenerPorID)
			cotizaciones.GET("/numero/:numero", ctrl.ObtenerPorNumero)
			cotizaciones.PUT("/:id", ctrl.Actualizar)
			cotizaciones.PUT("/:id/estado", ctrl.ActualizarEstado)
			cotizaciones.DELETE("/:id", ctrl.Eliminar)
		}
	}
}

func (ctrl *CotizacionController) ObtenerTodas(gctx *gin.Context) {
	// Parámetros de paginación
	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Filtros
	filters := make(map[string]interface{})
	if estado := gctx.Query("estado"); estado != "" {
		filters["estado"] = estado
	}
	if servicio := gctx.Query("servicio"); servicio != "" {
		filters["servicio"] = servicio
	}
	if prioridad := gctx.Query("prioridad"); prioridad != "" {
		filters["prioridad"] = prioridad
	}

	cotizaciones, total, err := ctrl.cotizacionFacade.ObtenerTodas(gctx.Request.Context(), page, limit, filters)
	if err != nil {
		gctx.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cotizaciones,
		Pagination: &types.PaginationResponse{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   int(total),
			ItemsPerPage: limit,
		},
	})
}

func (ctrl *CotizacionController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	cotizacion, err := ctrl.cotizacionFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cotizacion,
	})
}

func (ctrl *CotizacionController) ObtenerPorNumero(gctx *gin.Context) {
	numero := gctx.Param("numero")

	cotizacion, err := ctrl.cotizacionFacade.ObtenerPorNumero(gctx.Request.Context(), numero)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cotizacion,
	})
}

func (ctrl *CotizacionController) Crear(gctx *gin.Context) {
	var r recipe.CrearCotizacionRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	cotizacion, err := ctrl.cotizacionFacade.Crear(
		gctx.Request.Context(),
		r.Nombre,
		r.Email,
		r.Empresa,
		r.Telefono,
		r.Servicio,
		r.Plazo,
		r.Mensaje,
	)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data:    cotizacion,
		Message: "Cotización creada correctamente",
	})
}

func (ctrl *CotizacionController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarCotizacionRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	updates := make(map[string]interface{})
	if r.Nombre != "" {
		updates["nombre"] = r.Nombre
	}
	if r.Email != "" {
		updates["email"] = r.Email
	}
	if r.Empresa != "" {
		updates["empresa"] = r.Empresa
	}
	if r.Telefono != "" {
		updates["telefono"] = r.Telefono
	}
	if r.Estado != "" {
		updates["estado"] = r.Estado
	}
	if r.Prioridad != "" {
		updates["prioridad"] = r.Prioridad
	}

	cotizacion, err := ctrl.cotizacionFacade.Actualizar(gctx.Request.Context(), id, updates)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cotizacion,
		Message: "Cotización actualizada correctamente",
	})
}

func (ctrl *CotizacionController) ActualizarEstado(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarEstadoCotizacionRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.cotizacionFacade.ActualizarEstado(gctx.Request.Context(), id, r.Estado)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Estado actualizado correctamente",
	})
}

func (ctrl *CotizacionController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.cotizacionFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Cotización eliminada correctamente",
	})
}
