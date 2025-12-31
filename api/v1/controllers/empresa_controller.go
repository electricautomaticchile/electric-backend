package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmpresaController struct {
	empresaFacade *facades.EmpresaFacade
}

func NewEmpresaController(empresaFacade *facades.EmpresaFacade) *EmpresaController {
	return &EmpresaController{
		empresaFacade: empresaFacade,
	}
}

func (ctrl *EmpresaController) ObtenerTodas(gctx *gin.Context) {
	empresas, err := ctrl.empresaFacade.ObtenerTodas(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    empresas,
	})
}

func (ctrl *EmpresaController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	empresa, err := ctrl.empresaFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    empresa,
	})
}

func (ctrl *EmpresaController) Crear(gctx *gin.Context) {
	var r recipe.CrearEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	empresa, err := ctrl.empresaFacade.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    empresa,
		"message": "Empresa creada correctamente",
	})
}

func (ctrl *EmpresaController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	empresa, err := ctrl.empresaFacade.Actualizar(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    empresa,
		"message": "Empresa actualizada correctamente",
	})
}

func (ctrl *EmpresaController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.empresaFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Empresa eliminada correctamente",
	})
}

func (ctrl *EmpresaController) CambiarEstado(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.CambiarEstadoEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.empresaFacade.CambiarEstado(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Estado actualizado correctamente",
	})
}
