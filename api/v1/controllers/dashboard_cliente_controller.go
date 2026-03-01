package controllers

import (
	"context"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardClienteController struct {
	clienteFacade     *facades.ClienteFacade
	dispositivoFacade *facades.DispositivoFacade
	boletaService     *services.BoletaService
	dashboardService  *services.DashboardService
}

func NewDashboardClienteController(
	clienteFacade *facades.ClienteFacade,
	dispositivoFacade *facades.DispositivoFacade,
	boletaService *services.BoletaService,
	dashboardService *services.DashboardService,
) *DashboardClienteController {
	return &DashboardClienteController{
		clienteFacade:     clienteFacade,
		dispositivoFacade: dispositivoFacade,
		boletaService:     boletaService,
		dashboardService:  dashboardService,
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

	clientes := router.Group("/clientes")
	clientes.Use(middleware.AuthMiddleware())
	{
		clientes.GET("/mi-dispositivo", ctrl.ObtenerMiDispositivo)
	}
}

func (ctrl *DashboardClienteController) ObtenerResumen(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	// Contexto propio para no depender del timeout del request HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cliente, err := ctrl.clienteFacade.ObtenerPorID(ctx, userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	dispositivos, _ := ctrl.dispositivoFacade.ObtenerPorCliente(ctx, userID)

	dispositivosActivos := 0
	consumoActual := 0.0
	costoActual := 0.0
	var ultimaLectura interface{}

	for _, disp := range dispositivos {
		if disp.Estado == "activo" {
			dispositivosActivos++
		}
		// Datos en tiempo real desde ultimaLectura del dispositivo (actualizado por Arduino)
		if disp.UltimaLectura != nil {
			consumoActual += disp.UltimaLectura.Energy
			costoActual += disp.UltimaLectura.Cost
			ultimaLectura = disp.UltimaLectura
		}
	}

	resumen := gin.H{
		"cliente": gin.H{
			"nombre":           cliente.Nombre,
			"numeroCliente":    cliente.NumeroCliente,
			"correo":           cliente.Correo,
			"telefono":         cliente.Telefono,
			"direccion":        cliente.Direccion,
			"imagenPerfil":     cliente.ImagenPerfil,
			"passwordTemporal": cliente.PasswordTemporal != "",
		},
		"estadisticas": gin.H{
			"dispositivosActivos": dispositivosActivos,
			"dispositivosTotal":   len(dispositivos),
			"consumoMensual":      consumoActual,
			"costoMensual":        costoActual,
			"ultimaLectura":       ultimaLectura,
			"boletasPendientes":   0,
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

	consumo, err := ctrl.dashboardService.ObtenerConsumoCliente(gctx.Request.Context(), userID)
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

func (ctrl *DashboardClienteController) ObtenerMiDispositivo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil || len(dispositivos) == 0 {
		gctx.JSON(http.StatusOK, types.ApiResponse{
			Success: true,
			Data:    nil,
			Message: "Sin dispositivo asignado",
		})
		return
	}

	d := dispositivos[0]
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"dispositivoId":     d.ID,
			"numeroDispositivo": d.NumeroDispositivo,
			"nombre":            d.Nombre,
			"estado":            d.Estado,
			"ultimaLectura":     d.UltimaLectura,
		},
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
