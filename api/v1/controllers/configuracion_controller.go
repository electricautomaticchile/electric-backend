package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfiguracionController struct {
	configuracionService *services.ConfiguracionService
}

func NewConfiguracionController(configuracionService *services.ConfiguracionService) *ConfiguracionController {
	return &ConfiguracionController{
		configuracionService: configuracionService,
	}
}

func (ctrl *ConfiguracionController) ObtenerTodas(gctx *gin.Context) {
	configuraciones, err := ctrl.configuracionService.ObtenerTodas(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configuraciones,
	})
}

func (ctrl *ConfiguracionController) ObtenerPorClave(gctx *gin.Context) {
	clave := gctx.Param("clave")

	configuracion, err := ctrl.configuracionService.ObtenerPorClave(gctx.Request.Context(), clave)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configuracion,
	})
}

func (ctrl *ConfiguracionController) Actualizar(gctx *gin.Context) {
	var r recipe.ActualizarConfiguracionRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.configuracionService.Actualizar(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Configuración actualizada correctamente",
})
}
