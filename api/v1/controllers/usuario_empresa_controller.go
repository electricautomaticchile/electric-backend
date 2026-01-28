package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"
	"strconv"

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

	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	
	skip := (page - 1) * limit

	usuarios, err := ctrl.usuarioService.ObtenerTodos(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	total := len(usuarios)
	end := skip + limit
	if end > total {
		end = total
	}
	
	paginatedUsuarios := usuarios
	if skip < total {
		paginatedUsuarios = usuarios[skip:end]
	} else {
		paginatedUsuarios = []models.UsuarioEmpresaModel{}
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    paginatedUsuarios,
		Pagination: &types.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
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
