package controllers

import (
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ImagenPerfilController struct {
	service *services.ImagenPerfilService
}

func NewImagenPerfilController(service *services.ImagenPerfilService) *ImagenPerfilController {
	return &ImagenPerfilController{
		service: service,
	}
}

// SetupRoutes configura las rutas del controlador
func (ctrl *ImagenPerfilController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/imagenes-perfil")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.POST("/:tipoUsuario/:userId/upload", ctrl.SubirYActualizarImagen)
	g.GET("/:tipoUsuario/:userId", ctrl.ObtenerImagenPerfil)
	g.DELETE("/:tipoUsuario/:userId", ctrl.EliminarImagenPerfil)
}

func (ctrl *ImagenPerfilController) SubirYActualizarImagen(gctx *gin.Context) {
	tipoUsuario := gctx.Param("tipoUsuario")
	userID := gctx.Param("userId")

	file, header, err := gctx.Request.FormFile("imagen")
	if err != nil {
		gctx.Error(types.ThrowPower("No se proporcionó archivo de imagen"))
		return
	}
	defer file.Close()

	imageURL, err := ctrl.service.SubirImagenPerfil(file, header, tipoUsuario, userID)
	if err != nil {
		gctx.Error(types.ThrowPower("Error subiendo imagen: " + err.Error()))
		return
	}

	if err := ctrl.service.ActualizarImagenPerfil(imageURL, tipoUsuario, userID); err != nil {
		gctx.Error(types.ThrowPower("Error actualizando perfil: " + err.Error()))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"imageURL": imageURL,
		},
		Message: "Imagen de perfil actualizada correctamente",
	})
}

func (ctrl *ImagenPerfilController) ObtenerImagenPerfil(gctx *gin.Context) {
	tipoUsuario := gctx.Param("tipoUsuario")
	userID := gctx.Param("userId")

	imageURL, err := ctrl.service.ObtenerImagenPerfil(tipoUsuario, userID)
	if err != nil {
		gctx.Error(types.ThrowPower("Imagen no encontrada: " + err.Error()))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"imageURL": imageURL,
		},
	})
}

func (ctrl *ImagenPerfilController) EliminarImagenPerfil(gctx *gin.Context) {
	tipoUsuario := gctx.Param("tipoUsuario")
	userID := gctx.Param("userId")

	if err := ctrl.service.EliminarImagenPerfil(tipoUsuario, userID); err != nil {
		gctx.Error(types.ThrowPower("Error eliminando imagen: " + err.Error()))
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Message: "Imagen de perfil eliminada correctamente",
	})
}
