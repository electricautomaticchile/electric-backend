package controllers

import (
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardClienteController struct {
	clienteFacade      *facades.ClienteFacade
	dispositivoFacade  *facades.DispositivoFacade
	estadisticaService interface {
		ObtenerConsumoCliente(clienteID string, fechaInicio, fechaFin string) (interface{}, error)
	}
}

func NewDashboardClienteController(
	clienteFacade *facades.ClienteFacade,
	dispositivoFacade *facades.DispositivoFacade,
) *DashboardClienteController {
	return &DashboardClienteController{
		clienteFacade:     clienteFacade,
		dispositivoFacade: dispositivoFacade,
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
	}
}

func (ctrl *DashboardClienteController) ObtenerResumen(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}
	
	resumen := gin.H{
		"cliente": gin.H{
			"nombre":        cliente.Nombre,
			"numeroCliente": cliente.NumeroCliente,
			"correo":        cliente.Correo,
		},
		"estadisticas": gin.H{
			"dispositivosActivos": 0,
			"consumoMensual":      0,
			"costoMensual":        0,
		},
	}
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resumen,
	})
}

func (ctrl *DashboardClienteController) ObtenerDispositivos(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"dispositivos": []interface{}{},
			"clienteId":    userID,
		},
	})
}

func (ctrl *DashboardClienteController) ObtenerConsumo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)
	
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"consumo":   []interface{}{},
			"clienteId": userID,
		},
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
