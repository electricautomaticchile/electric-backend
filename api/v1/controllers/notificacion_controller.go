package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificacionController struct {
	notificacionService *services.NotificacionService
}

func NewNotificacionController(notificacionService *services.NotificacionService) *NotificacionController {
	return &NotificacionController{
		notificacionService: notificacionService,
	}
}

func (ctrl *NotificacionController) Listar(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	notificaciones, err := ctrl.notificacionService.Listar(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notificaciones,
	})
}

func (ctrl *NotificacionController) MarcarLeida(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.notificacionService.MarcarLeida(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Notificación marcada como leída",
})
}

func (ctrl *NotificacionController) MarcarTodasLeidas(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	err := ctrl.notificacionService.MarcarTodasLeidas(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Todas las notificaciones marcadas como leídas",
})
}

func (ctrl *NotificacionController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.notificacionService.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
"success": true,
"message": "Notificación eliminada",
})
}

func (ctrl *NotificacionController) ObtenerEstadisticas(gctx *gin.Context) {
	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":    0,
			"noLeidas": 0,
			"leidas":   0,
		},
	})
}
