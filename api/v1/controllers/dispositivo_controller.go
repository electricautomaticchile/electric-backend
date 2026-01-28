package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/domain/models"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DispositivoController struct {
	dispositivoFacade *facades.DispositivoFacade
}

func NewDispositivoController(dispositivoFacade *facades.DispositivoFacade) *DispositivoController {
	return &DispositivoController{
		dispositivoFacade: dispositivoFacade,
	}
}

func (ctrl *DispositivoController) ObtenerTodos(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
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

	dispositivos, err := ctrl.dispositivoFacade.ObtenerTodos(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}
	
	total := len(dispositivos)
	end := skip + limit
	if end > total {
		end = total
	}
	
	paginatedDispositivos := dispositivos
	if skip < total {
		paginatedDispositivos = dispositivos[skip:end]
	} else {
		paginatedDispositivos = []models.DispositivoModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedDispositivos,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (ctrl *DispositivoController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivo, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
	})
}

func (ctrl *DispositivoController) ObtenerPorCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivos,
	})
}

func (ctrl *DispositivoController) Crear(gctx *gin.Context) {
	var r recipe.CrearDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo creado correctamente",
	})
}

func (ctrl *DispositivoController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.Actualizar(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo actualizado correctamente",
	})
}

func (ctrl *DispositivoController) AsignarCliente(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.AsignarDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.AsignarCliente(gctx.Request.Context(), id, r.ClienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo asignado correctamente",
	})
}

func (ctrl *DispositivoController) DesasignarCliente(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivo, err := ctrl.dispositivoFacade.AsignarCliente(gctx.Request.Context(), id, "")
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo desasignado correctamente",
	})
}

func (ctrl *DispositivoController) ActualizarLectura(gctx *gin.Context) {
	numeroDispositivo := gctx.Param("numeroDispositivo")

	var r recipe.ActualizarLecturaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.dispositivoFacade.ActualizarUltimaLectura(gctx.Request.Context(), numeroDispositivo, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Lectura actualizada correctamente",
})
}

func (ctrl *DispositivoController) CambiarEstado(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.CambiarEstadoDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.dispositivoFacade.CambiarEstado(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Estado actualizado correctamente",
})
}

func (ctrl *DispositivoController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.dispositivoFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Dispositivo eliminado correctamente",
})
}
