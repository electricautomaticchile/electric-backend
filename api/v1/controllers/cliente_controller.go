package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"net/http"
	"strconv"

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
	clientes.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
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

	paginated := gctx.Query("paginated")
	if paginated == "true" {
		ctrl.ObtenerTodosPaginado(gctx)
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

func (ctrl *ClienteController) ObtenerTodosPaginado(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	page := 1
	pageSize := 10
	sortBy := gctx.DefaultQuery("sortBy", "fechaCreacion")
	sortDir := gctx.DefaultQuery("sortDir", "desc")

	if p := gctx.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	if ps := gctx.Query("pageSize"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil {
			pageSize = val
		}
	}

	filters := types.FilterParams{
		Search:   gctx.Query("search"),
		DateFrom: gctx.Query("dateFrom"),
		DateTo:   gctx.Query("dateTo"),
		Active:   types.ParseBoolPtr(gctx.Query("active")),
		Status:   gctx.Query("status"),
		Type:     gctx.Query("type"),
	}

	params := types.NewPaginationParams(page, pageSize, sortBy, sortDir)
	clientes, total, err := ctrl.clienteFacade.ObtenerTodosPaginado(gctx.Request.Context(), empresaID.(string), params, filters)
	if err != nil {
		gctx.Error(err)
		return
	}

	response := types.NewPaginatedResponse(clientes, page, pageSize, total)
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    response,
	})
}

func (ctrl *ClienteController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, cliente.ID, cliente.EmpresaID, true) {
		gctx.Error(types.ThrowPower("No tienes acceso a este cliente"))
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
	if !middleware.CanAccessResource(gctx, cliente.ID, cliente.EmpresaID, true) {
		gctx.Error(types.ThrowPower("No tienes acceso a este cliente"))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
	})
}

func (ctrl *ClienteController) Crear(gctx *gin.Context) {
	if !middleware.IsEmpresaContext(gctx) || middleware.EmpresaID(gctx) == "" {
		gctx.Error(types.ThrowPower("Solo una empresa autenticada puede crear clientes"))
		return
	}

	var r recipe.CrearClienteRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}
	r.EmpresaID = middleware.EmpresaID(gctx)

	if err := validation.ValidateNombre(r.Nombre); err != nil {
		gctx.Error(types.ThrowRecipe(err.Error(), ""))
		return
	}

	if r.Rut != "" {
		if !validation.ValidarRUT(r.Rut) {
			gctx.Error(types.ThrowRecipe("RUT inválido", ""))
			return
		}
	}

	if err := validation.ValidateEmail(r.Correo); err != nil {
		gctx.Error(types.ThrowRecipe(err.Error(), ""))
		return
	}

	if r.Direccion != "" {
		if err := validation.ValidateDireccion(r.Direccion); err != nil {
			gctx.Error(types.ThrowRecipe(err.Error(), ""))
			return
		}
	}

	if r.Ciudad != "" {
		if err := validation.ValidateComuna(r.Ciudad); err != nil {
			gctx.Error(types.ThrowRecipe(err.Error(), ""))
			return
		}
	}

	if r.Telefono != "" {
		if !validation.ValidarTelefonoChileno(r.Telefono) {
			gctx.Error(types.ThrowRecipe("Teléfono inválido", ""))
			return
		}
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

	clienteActual, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, clienteActual.ID, clienteActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para actualizar este cliente"))
		return
	}

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

	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, cliente.ID, cliente.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para eliminar este cliente"))
		return
	}

	err = ctrl.clienteFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Cliente eliminado correctamente",
	})
}
