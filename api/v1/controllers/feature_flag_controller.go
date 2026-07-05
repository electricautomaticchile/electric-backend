package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeatureFlagController struct {
	featureFlagService *services.FeatureFlagService
}

func NewFeatureFlagController(featureFlagService *services.FeatureFlagService) *FeatureFlagController {
	return &FeatureFlagController{
		featureFlagService: featureFlagService,
	}
}

// SetupRoutes expone la administración de flags bajo /feature-flags.
// Solo administradores globales pueden listar/modificar/eliminar flags.
func (ctrl *FeatureFlagController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/feature-flags")
	g.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin", "superadmin", "super_admin"))
	g.GET("", ctrl.Listar)
	g.PUT("", middleware.CSRFMiddleware(), ctrl.Set)
	g.DELETE("/:key", middleware.CSRFMiddleware(), ctrl.Eliminar)
}

func (ctrl *FeatureFlagController) Listar(gctx *gin.Context) {
	flags, err := ctrl.featureFlagService.Listar(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}
	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Data: flags})
}

func (ctrl *FeatureFlagController) Set(gctx *gin.Context) {
	var body recipe.SetFeatureFlagRecipe
	if err := gctx.ShouldBindJSON(&body); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	if err := ctrl.featureFlagService.Set(
		gctx.Request.Context(), body.Key, body.Descripcion, body.Enabled, body.EmpresaIDs,
	); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Message: "Flag actualizado"})
}

func (ctrl *FeatureFlagController) Eliminar(gctx *gin.Context) {
	key := gctx.Param("key")
	if err := ctrl.featureFlagService.Eliminar(gctx.Request.Context(), key); err != nil {
		gctx.Error(err)
		return
	}
	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Message: "Flag eliminado"})
}
