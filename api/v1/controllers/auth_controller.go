package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authFacade *facades.AuthFacade
}

func NewAuthController(authFacade *facades.AuthFacade) *AuthController {
	return &AuthController{
		authFacade: authFacade,
	}
}

// SetupRoutes configura las rutas del controlador
func (ctrl *AuthController) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", ctrl.Login)
		auth.POST("/login/empresa", ctrl.LoginEmpresa)
		auth.POST("/registro-empresa", ctrl.RegistroEmpresa)
		auth.POST("/solicitar-recuperacion", ctrl.SolicitarRecuperacion)
		auth.POST("/restablecer-password", ctrl.RestablecerPassword)

		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/profile", ctrl.ObtenerPerfil)
			auth.POST("/cambiar-password", ctrl.CambiarPassword)
			auth.POST("/logout", ctrl.Logout)
		}
	}
}

func (ctrl *AuthController) Login(gctx *gin.Context) {
	var r recipe.LoginRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	result, err := ctrl.authFacade.Login(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	domain := ""
	secure := false
	if gctx.Request.Host != "localhost:4000" && gctx.Request.Host != "127.0.0.1:4000" {
		domain = "api-electricautomaticchile.com"
		secure = true
	}

	gctx.SetSameSite(http.SameSiteNoneMode)
	gctx.SetCookie(
		"auth_token",
		result.Token,
		86400,
		"/",
		domain,
		secure,
		false,
	)

	requiereCambioPassword := result.User.PasswordTemporal != ""
	
	if requiereCambioPassword {
		gctx.SetCookie(
			"requiereCambioPassword",
			"true",
			86400,
			"/",
			domain,
			secure,
			false,
		)
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"token":                  result.Token,
			"refreshToken":           result.RefreshToken,
			"user":                   result.User,
			"requiereCambioPassword": requiereCambioPassword,
		},
	})
}

func (ctrl *AuthController) Logout(gctx *gin.Context) {
	// Eliminar cookie
	gctx.SetCookie(
		"auth_token",
		"",
		-1,     // maxAge negativo elimina la cookie
		"/",
		"",
		false,
		true,
	)

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Sesión cerrada correctamente",
	})
}

func (ctrl *AuthController) ObtenerPerfil(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	usuario, err := ctrl.authFacade.ObtenerPerfil(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    usuario,
	})
}

func (ctrl *AuthController) CambiarPassword(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	var r recipe.CambiarPasswordRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.authFacade.CambiarPassword(gctx.Request.Context(), userID, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.SetCookie(
		"requiereCambioPassword",
		"",
		-1,
		"/",
		"",
		false,
		false,
	)

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Contraseña actualizada correctamente",
	})
}

func (ctrl *AuthController) SolicitarRecuperacion(gctx *gin.Context) {
	var r recipe.SolicitarRecuperacionRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.authFacade.SolicitarRecuperacion(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Si el email existe, recibirás instrucciones para recuperar tu contraseña",
	})
}

func (ctrl *AuthController) RestablecerPassword(gctx *gin.Context) {
	var r recipe.RestablecerPasswordRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.authFacade.RestablecerPassword(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Contraseña restablecida correctamente",
	})
}

func (ctrl *AuthController) RegistroEmpresa(gctx *gin.Context) {
	var r recipe.RegistroEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	empresa, err := ctrl.authFacade.RegistrarEmpresa(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"empresa":         empresa,
			"numeroCliente":   empresa.NumeroCliente,
			"passwordTemporal": empresa.Password,
		},
		Message: "Empresa registrada correctamente. Guarda tu número de cliente y contraseña temporal",
	})
}

func (ctrl *AuthController) LoginEmpresa(gctx *gin.Context) {
	var r recipe.LoginEmpresaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	result, err := ctrl.authFacade.LoginEmpresa(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	domain := ""
	secure := false
	if gctx.Request.Host != "localhost:4000" && gctx.Request.Host != "127.0.0.1:4000" {
		domain = "api-electricautomaticchile.com"
		secure = true
	}

	gctx.SetSameSite(http.SameSiteNoneMode)
	gctx.SetCookie(
		"auth_token",
		result.Token,
		86400,
		"/",
		domain,
		secure,
		false,
	)

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"token":        result.Token,
			"refreshToken": result.RefreshToken,
			"user":         result.User,
			"permisos":     result.Permisos,
		},
	})
}
