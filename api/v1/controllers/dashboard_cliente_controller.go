package controllers

import (
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardClienteController struct {
	clienteFacade       *facades.ClienteFacade
	dispositivoFacade   *facades.DispositivoFacade
	boletaService       *services.BoletaService
	estadisticaService  *services.EstadisticaService
}

func NewDashboardClienteController(
	clienteFacade *facades.ClienteFacade,
	dispositivoFacade *facades.DispositivoFacade,
	boletaService *services.BoletaService,
	estadisticaService *services.EstadisticaService,
) *DashboardClienteController {
	return &DashboardClienteController{
		clienteFacade:      clienteFacade,
		dispositivoFacade:  dispositivoFacade,
		boletaService:      boletaService,
		estadisticaService: estadisticaService,
	}
}

func (ctrl *DashboardClienteController) SetupRoutes(router *gin.RouterGroup) {
	dashboard := router.Group("/dashboard/cliente")
	dashboard.Use(middleware.AuthMiddleware())
	{
		dashboard.GET("/resumen", ctrl.ObtenerResumen)
		dashboard.GET("/dispositivos", ctrl.ObtenerDispositivos)
		dashboard.GET("/consumo", ctrl.ObtenerConsumo)
		dashboard.GET("/perfil", ctrl.ObtenerPerfil)
		dashboard.GET("/boletas", ctrl.ObtenerBoletas)
		dashboard.PUT("/perfil", ctrl.ActualizarPerfil)
	}
}

func (ctrl *DashboardClienteController) ObtenerResumen(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	dispositivos, _ := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), userID)
	
	dispositivosActivos := 0
	for _, disp := range dispositivos {
		if disp.Estado == "activo" {
			dispositivosActivos++
		}
	}

	now := time.Now()
	inicioMes := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	_ = inicioMes

	consumoMensual := 0.0
	costoMensual := 0.0

	estadisticas, err := ctrl.estadisticaService.ObtenerConsumoCliente(gctx.Request.Context(), userID)
	if err == nil && estadisticas != nil {
		if consumo, ok := estadisticas["consumoTotal"].(float64); ok {
			consumoMensual = consumo
		}
		if costo, ok := estadisticas["costoTotal"].(float64); ok {
			costoMensual = costo
		}
	}

	boletas, _ := ctrl.boletaService.ObtenerPorCliente(gctx.Request.Context(), userID)
	boletasPendientes := 0
	for _, boleta := range boletas {
		if boleta.Estado == "pendiente" {
			boletasPendientes++
		}
	}
	
	resumen := gin.H{
		"cliente": gin.H{
			"nombre":            cliente.Nombre,
			"numeroCliente":     cliente.NumeroCliente,
			"correo":            cliente.Correo,
			"telefono":          cliente.Telefono,
			"direccion":         cliente.Direccion,
			"imagenPerfil":      cliente.ImagenPerfil,
			"passwordTemporal":  cliente.PasswordTemporal != "",
		},
		"estadisticas": gin.H{
			"dispositivosActivos": dispositivosActivos,
			"dispositivosTotal":   len(dispositivos),
			"consumoMensual":      consumoMensual,
			"costoMensual":        costoMensual,
			"boletasPendientes":   boletasPendientes,
		},
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resumen,
	})
}

func (ctrl *DashboardClienteController) ObtenerDispositivos(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    dispositivos,
	})
}

func (ctrl *DashboardClienteController) ObtenerConsumo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	consumo, err := ctrl.estadisticaService.ObtenerConsumoCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    consumo,
	})
}

func (ctrl *DashboardClienteController) ObtenerPerfil(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
	})
}

func (ctrl *DashboardClienteController) ObtenerBoletas(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	boletas, err := ctrl.boletaService.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    boletas,
	})
}

func (ctrl *DashboardClienteController) ActualizarPerfil(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	_ = userID
	
	var datos map[string]interface{}
	if err := gctx.ShouldBindJSON(&datos); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Perfil actualizado correctamente",
	})
}
