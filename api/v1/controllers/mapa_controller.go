package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MapaController struct {
	dispositivoService *services.DispositivoService
	clienteService     *services.ClienteService
}

func NewMapaController(dispositivoService *services.DispositivoService, clienteService *services.ClienteService) *MapaController {
	return &MapaController{
		dispositivoService: dispositivoService,
		clienteService:     clienteService,
	}
}

func (ctrl *MapaController) ObtenerDatosMapa(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	dispositivos, err := ctrl.dispositivoService.ObtenerDispositivosConUbicacion(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	clientes, err := ctrl.clienteService.ObtenerClientesConUbicacion(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success":      true,
		"dispositivos": dispositivos,
		"clientes":     clientes,
	})
}
