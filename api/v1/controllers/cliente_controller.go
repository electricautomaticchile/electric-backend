package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClienteController struct {
	clienteFacade *facades.ClienteFacade
}

func NewClienteController(clienteFacade *facades.ClienteFacade) *ClienteController {
	return &ClienteController{
		clienteFacade: clienteFacade,
	}
}

// SetupRoutes configura las rutas del controlador
func (ctrl *ClienteController) SetupRoutes(router *gin.RouterGroup) {
	clientes := router.Group("/clientes")
	clientes.Use(middleware.AuthMiddleware())
	{
		clientes.GET("", ctrl.ObtenerTodos)
		clientes.GET("/:id", ctrl.ObtenerPorID)
		clientes.GET("/numero/:numeroCliente", ctrl.ObtenerPorNumero)
		clientes.POST("", ctrl.Crear)
		clientes.PUT("/:id", ctrl.Actualizar)
		clientes.DELETE("/:id", ctrl.Eliminar)
	}
}

func (ctrl *ClienteController) ObtenerTodos(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	clientes, err := ctrl.clienteFacade.ObtenerTodos(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    clientes,
	})
}

func (ctrl *ClienteController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
	})
}

func (ctrl *ClienteController) ObtenerPorNumero(gctx *gin.Context) {
	numeroCliente := gctx.Param("numeroCliente")

	cliente, err := ctrl.clienteFacade.ObtenerPorNumero(gctx.Request.Context(), numeroCliente)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
	})
}

func (ctrl *ClienteController) Crear(gctx *gin.Context) {
	var r recipe.CrearClienteRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	cliente, err := ctrl.clienteFacade.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	passwordTemporal := cliente.PasswordTemporal
	cliente.PasswordTemporal = ""

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"cliente":          cliente,
			"numeroCliente":    cliente.NumeroCliente,
			"passwordTemporal": passwordTemporal,
		},
		Message: "Cliente creado correctamente. Guarda el número de cliente y contraseña temporal",
	})
}

func (ctrl *ClienteController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarClienteRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	cliente, err := ctrl.clienteFacade.Actualizar(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
		Message: "Cliente actualizado correctamente",
	})
}

func (ctrl *ClienteController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.clienteFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Cliente eliminado correctamente",
	})
}
