package middleware

import (
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateRUT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.Next()
			return
		}

		if rut, exists := body["rut"]; exists {
			rutStr, ok := rut.(string)
			if ok && rutStr != "" {
				if !validation.ValidarRUT(rutStr) {
					c.AbortWithStatusJSON(http.StatusBadRequest, types.ApiResponse{
						Success: false,
						Message: "RUT inválido",
					})
					return
				}
				body["rut"] = validation.NormalizarRUT(rutStr)
			}
		}

		c.Set("validatedBody", body)
		c.Next()
	}
}
