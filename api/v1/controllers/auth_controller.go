package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/config"
	"electric-backend/domain/facades"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"os"
	"strings"

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
		auth.GET("/csrf-token", ctrl.GetCSRFToken)
		auth.POST("/login", ctrl.Login)
		auth.POST("/login/empresa", ctrl.LoginEmpresa)
		auth.POST("/registro-empresa", ctrl.RegistroEmpresa)
		auth.POST("/solicitar-recuperacion", ctrl.SolicitarRecuperacion)
		auth.POST("/restablecer-password", ctrl.RestablecerPassword)
		auth.POST("/refresh", ctrl.RefreshToken)
		auth.POST("/refresh-token", ctrl.RefreshToken)

		auth.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
		{
			auth.GET("/profile", ctrl.ObtenerPerfil)
			auth.GET("/me", ctrl.ObtenerPerfil)
			auth.POST("/cambiar-password", ctrl.CambiarPassword)
			auth.POST("/logout", ctrl.Logout)
		}
	}
}

func authCookieDomain() string {
	if config.AppConfig == nil {
		return ""
	}
	return config.AppConfig.AuthCookieDomain
}

func secureCookies() bool {
	if config.AppConfig != nil && config.AppConfig.Environment == "production" {
		return true
	}
	return os.Getenv("NODE_ENV") == "production"
}

func setSessionCookies(gctx *gin.Context, token string, refreshToken string) {
	const authMaxAge = 24 * 60 * 60
	const refreshMaxAge = 7 * 24 * 60 * 60

	gctx.SetSameSite(http.SameSiteLaxMode)
	gctx.SetCookie("auth_token", token, authMaxAge, "/", authCookieDomain(), secureCookies(), true)
	gctx.SetCookie("refresh_token", refreshToken, refreshMaxAge, "/", authCookieDomain(), secureCookies(), true)
}

func clearSessionCookies(gctx *gin.Context) {
	gctx.SetSameSite(http.SameSiteLaxMode)
	gctx.SetCookie("auth_token", "", -1, "/", authCookieDomain(), secureCookies(), true)
	gctx.SetCookie("refresh_token", "", -1, "/", authCookieDomain(), secureCookies(), true)
	gctx.SetCookie("requiereCambioPassword", "", -1, "/", authCookieDomain(), secureCookies(), false)
}

func shouldExposeTokens(gctx *gin.Context) bool {
	clientType := strings.ToLower(gctx.GetHeader("X-Client-Type"))
	return clientType == "mobile" || clientType == "native"
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

	requiereCambioPassword := result.User.PasswordTemporal != ""

	setSessionCookies(gctx, result.Token, result.RefreshToken)

	data := gin.H{
		"user":                   result.User,
		"requiereCambioPassword": requiereCambioPassword,
	}
	if shouldExposeTokens(gctx) {
		data["token"] = result.Token
		data["refreshToken"] = result.RefreshToken
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    data,
	})
}

func (ctrl *AuthController) Logout(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	ctrl.authFacade.RevokeAllRefreshTokens(gctx.Request.Context(), userID)
	clearSessionCookies(gctx)

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Sesión cerrada correctamente",
	})
}

func (ctrl *AuthController) GetCSRFToken(gctx *gin.Context) {
	sessionID := gctx.GetString("session_id")
	if sessionID == "" {
		userID, exists := gctx.Get("userID")
		if exists {
			sessionID = userID.(string)
		}
	}
	if sessionID == "" {
		if cookieToken, err := gctx.Cookie("auth_token"); err == nil && cookieToken != "" {
			if claims, err := middleware.ParseJWTClaims(cookieToken); err == nil {
				sessionID = claims.UserID
			}
		}
	}
	if sessionID == "" {
		if userID := gctx.GetString("userId"); userID != "" {
			sessionID = userID
		} else {
			sessionID = gctx.ClientIP()
		}
	}

	token, err := middleware.GenerateCSRFTokenForSession(sessionID)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, types.ApiResponse{
			Success: false,
			Message: "Error generando token CSRF",
		})
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"token": token,
		},
	})
}

func (ctrl *AuthController) RefreshToken(gctx *gin.Context) {
	var r struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := gctx.ShouldBindJSON(&r); err != nil {
		r.RefreshToken = ""
	}

	if r.RefreshToken == "" {
		if cookieToken, err := gctx.Cookie("refresh_token"); err == nil {
			r.RefreshToken = cookieToken
		}
	}

	if r.RefreshToken == "" {
		gctx.Error(types.ThrowRecipe("Token requerido", "refreshToken"))
		return
	}

	result, err := ctrl.authFacade.RefreshToken(gctx.Request.Context(), r.RefreshToken)
	if err != nil {
		gctx.Error(err)
		return
	}

	setSessionCookies(gctx, result.Token, result.RefreshToken)

	data := gin.H{}
	if shouldExposeTokens(gctx) {
		data["token"] = result.Token
		data["refreshToken"] = result.RefreshToken
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    data,
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

	gctx.SetCookie("requiereCambioPassword", "", -1, "/", authCookieDomain(), secureCookies(), false)

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
			"empresa":          empresa,
			"numeroCliente":    empresa.NumeroCliente,
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

	setSessionCookies(gctx, result.Token, result.RefreshToken)

	data := gin.H{
		"user":     result.User,
		"permisos": result.Permisos,
	}
	if shouldExposeTokens(gctx) {
		data["token"] = result.Token
		data["refreshToken"] = result.RefreshToken
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    data,
	})
}
