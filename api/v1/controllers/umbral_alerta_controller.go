package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UmbralAlertaController struct {
	umbralService *services.UmbralService
}

func NewUmbralAlertaController(umbralService *services.UmbralService) *UmbralAlertaController {
	return &UmbralAlertaController{
		umbralService: umbralService,
	}
}

// SetupRoutes expone la gestión de umbrales de alerta bajo /umbrales-alerta.
// Cada empresa gestiona sus propios umbrales; también pueden hacerlo los admins.
func (ctrl *UmbralAlertaController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/umbrales-alerta")
	g.Use(middleware.AuthMiddleware(), middleware.RequireRole("empresa", "admin", "superadmin", "super_admin"))
	g.GET("", ctrl.Obtener)
	g.PUT("", middleware.CSRFMiddleware(), ctrl.Set)
}

// empresaIDDesdeContexto extrae el empresaID del contexto de la petición.
func empresaIDDesdeContexto(gctx *gin.Context) (string, bool) {
	v := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	if !ok || id == "" {
		return "", false
	}
	return id, true
}

func (ctrl *UmbralAlertaController) Obtener(gctx *gin.Context) {
	empresaID, ok := empresaIDDesdeContexto(gctx)
	if !ok {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	umbral := ctrl.umbralService.ObtenerModelo(gctx.Request.Context(), empresaID)
	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Data: umbral})
}

func (ctrl *UmbralAlertaController) Set(gctx *gin.Context) {
	empresaID, ok := empresaIDDesdeContexto(gctx)
	if !ok {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	var body recipe.SetUmbralAlertaRecipe
	if err := gctx.ShouldBindJSON(&body); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}
	if body.VoltajeMin >= body.VoltajeMax {
		gctx.Error(types.ThrowRecipe("voltajeMin debe ser menor que voltajeMax", ""))
		return
	}

	if err := ctrl.umbralService.Guardar(
		gctx.Request.Context(), empresaID,
		body.VoltajeMin, body.VoltajeMax, body.CorrienteMax, body.ConsumoMax,
	); err != nil {
		gctx.Error(err)
		return
	}

	umbral := ctrl.umbralService.ObtenerModelo(gctx.Request.Context(), empresaID)
	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Message: "Umbrales actualizados", Data: umbral})
}
