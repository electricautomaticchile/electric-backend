package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsuarioEmpresaController struct {
	usuarioService *services.UsuarioEmpresaService
}

func NewUsuarioEmpresaController(usuarioService *services.UsuarioEmpresaService) *UsuarioEmpresaController {
	return &UsuarioEmpresaController{
		usuarioService: usuarioService,
	}
}

func (ctrl *UsuarioEmpresaController) ObtenerTodos(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.JSON(http.StatusUnauthorized, types.ApiResponse{
			Success: false,
			Error:   "No autorizado",
		})
		return
	}

	usuarios, err := ctrl.usuarioService.ObtenerTodos(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    usuarios,
	})
}

func (ctrl *UsuarioEmpresaController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	usuario, err := ctrl.usuarioService.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.JSON(http.StatusNotFound, types.ApiResponse{
			Success: false,
			Error:   "Usuario no encontrado",
		})
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    usuario,
	})
}

func (ctrl *UsuarioEmpresaController) Crear(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.JSON(http.StatusUnauthorized, types.ApiResponse{
			Success: false,
			Error:   "No autorizado",
		})
		return
	}

	var r recipe.CrearUsuarioEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	usuario, err := ctrl.usuarioService.Crear(gctx.Request.Context(), empresaID.(string), &r)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data:    usuario,
	})
}

func (ctrl *UsuarioEmpresaController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarUsuarioEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.JSON(http.StatusBadRequest, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	usuario, err := ctrl.usuarioService.Actualizar(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    usuario,
	})
}

func (ctrl *UsuarioEmpresaController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	if err := ctrl.usuarioService.Eliminar(gctx.Request.Context(), id); err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    "Usuario eliminado correctamente",
	})
}
