package types

import (
	"fmt"
	"net/http"
)

// ApiResponse es el formato estándar de respuesta de la API
type ApiResponse struct {
	Success    bool                `json:"success"`
	Data       interface{}         `json:"data,omitempty"`
	Message    string              `json:"message,omitempty"`
	Error      string              `json:"error,omitempty"`
	Errors     []string            `json:"errors,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// PaginationResponse contiene información de paginación
type PaginationResponse struct {
	CurrentPage  int `json:"currentPage"`
	TotalPages   int `json:"totalPages"`
	TotalItems   int `json:"totalItems"`
	ItemsPerPage int `json:"itemsPerPage"`
}

// ErrorType representa el tipo de error
type ErrorType string

const (
	ErrorTypeData   ErrorType = "DATA_ERROR"   // Error de datos (422)
	ErrorTypePower  ErrorType = "POWER_ERROR"  // Error de permisos (403)
	ErrorTypeRecipe ErrorType = "RECIPE_ERROR" // Error de validación (400)
	ErrorTypeAuth   ErrorType = "AUTH_ERROR"   // Error de autenticación (401)
	ErrorTypeServer ErrorType = "SERVER_ERROR" // Error del servidor (500)
)

// AppError representa un error de la aplicación
type AppError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Field      string    `json:"field,omitempty"`
	StatusCode int       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (campo: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ThrowData crea un error de datos (422)
func ThrowData(message string) error {
	return &AppError{
		Type:       ErrorTypeData,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// ThrowPower crea un error de permisos (403)
func ThrowPower(message string) error {
	return &AppError{
		Type:       ErrorTypePower,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// ThrowRecipe crea un error de validación (400)
func ThrowRecipe(message string, field string) error {
	return &AppError{
		Type:       ErrorTypeRecipe,
		Message:    message,
		Field:      field,
		StatusCode: http.StatusBadRequest,
	}
}

// ThrowAuth crea un error de autenticación (401)
func ThrowAuth(message string) error {
	return &AppError{
		Type:       ErrorTypeAuth,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// ThrowServer crea un error del servidor (500)
func ThrowServer(message string) error {
	return &AppError{
		Type:       ErrorTypeServer,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// GetStatusCode obtiene el código de estado HTTP del error
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
