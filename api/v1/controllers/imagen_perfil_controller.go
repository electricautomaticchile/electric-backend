package controllers

import (
	"electric-backend/domain/services"
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

func (c *ImagenPerfilController) SubirImagen(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo leer el archivo"})
		return
	}
	defer file.Close()

	tipoUsuario := ctx.PostForm("tipoUsuario")
	userID := ctx.PostForm("userId")

	if tipoUsuario == "" || userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "tipoUsuario y userId son requeridos"})
		return
	}

	imageURL, err := c.service.SubirImagenPerfil(file, header, tipoUsuario, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"imageUrl": imageURL,
		"fileName": header.Filename,
	})
}

func (c *ImagenPerfilController) ActualizarImagenPerfil(ctx *gin.Context) {
	var req struct {
		ImageURL    string `json:"imageUrl" binding:"required"`
		TipoUsuario string `json:"tipoUsuario" binding:"required"`
		UserID      string `json:"userId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.ActualizarImagenPerfil(req.ImageURL, req.TipoUsuario, req.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Imagen de perfil actualizada exitosamente",
	})
}

func (c *ImagenPerfilController) ObtenerImagenPerfil(ctx *gin.Context) {
	tipoUsuario := ctx.Param("tipoUsuario")
	userID := ctx.Param("userId")

	imageURL, err := c.service.ObtenerImagenPerfil(tipoUsuario, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":      true,
		"imagenPerfil": imageURL,
	})
}

func (c *ImagenPerfilController) EliminarImagenPerfil(ctx *gin.Context) {
	tipoUsuario := ctx.Param("tipoUsuario")
	userID := ctx.Param("userId")

	if err := c.service.EliminarImagenPerfil(tipoUsuario, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Imagen de perfil eliminada exitosamente",
	})
}

func (c *ImagenPerfilController) SubirYActualizarImagen(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo leer el archivo"})
		return
	}
	defer file.Close()

	tipoUsuario := ctx.PostForm("tipoUsuario")
	userID := ctx.PostForm("userId")

	if tipoUsuario == "" || userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "tipoUsuario y userId son requeridos"})
		return
	}

	imageURL, err := c.service.SubirImagenPerfil(file, header, tipoUsuario, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.ActualizarImagenPerfil(imageURL, tipoUsuario, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"imageUrl": imageURL,
		"fileName": header.Filename,
		"message":  "Imagen de perfil actualizada exitosamente",
	})
}
