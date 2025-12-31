package middleware

import (
	"electric-backend/types"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler maneja los errores de la aplicación
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*types.AppError); ok {
				c.JSON(appErr.StatusCode, types.ApiResponse{
					Success: false,
					Error:   appErr.Message,
					Errors:  []string{appErr.Message},
				})
				return
			}

			log.Printf("Error no manejado: %v", err)
			c.JSON(http.StatusInternalServerError, types.ApiResponse{
				Success: false,
				Error:   "Error interno del servidor",
				Errors:  []string{err.Error()},
			})
		}
	}
}
