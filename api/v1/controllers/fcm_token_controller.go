package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FCMTokenController struct {
	fcmTokenService *services.FCMTokenService
}

func NewFCMTokenController(fcmTokenService *services.FCMTokenService) *FCMTokenController {
	return &FCMTokenController{
		fcmTokenService: fcmTokenService,
	}
}

// SetupRoutes configura las rutas del controlador.
// Se agrupa bajo /notificaciones para exponer POST /notificaciones/fcm-token,
// que es el endpoint que ya consume la app móvil.
func (ctrl *FCMTokenController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/notificaciones")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.POST("/fcm-token", ctrl.RegistrarToken)
}

func (ctrl *FCMTokenController) RegistrarToken(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	var body recipe.RegistrarFCMTokenRecipe
	if err := gctx.ShouldBindJSON(&body); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	if err := ctrl.fcmTokenService.RegistrarToken(gctx.Request.Context(), userID, body.Token, body.Plataforma); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Token FCM registrado",
	})
}
