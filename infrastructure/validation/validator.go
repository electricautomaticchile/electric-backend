package validation

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	validate.RegisterValidation("rut", validarRUTTag)
	validate.RegisterValidation("password_strong", validarPasswordTag)
	validate.RegisterValidation("telefono_cl", validarTelefonoTag)
	validate.RegisterValidation("periodo", validarPeriodoTag)
	validate.RegisterValidation("numero_cliente", validarNumeroClienteTag)
}

func validarRUTTag(fl validator.FieldLevel) bool {
	rut := fl.Field().String()
	return ValidarRUT(rut)
}

func validarPasswordTag(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return ValidarPassword(password) == nil
}

func validarTelefonoTag(fl validator.FieldLevel) bool {
	telefono := fl.Field().String()
	return ValidarTelefonoChileno(telefono)
}

func validarPeriodoTag(fl validator.FieldLevel) bool {
	periodo := fl.Field().String()
	return ValidarPeriodo(periodo)
}

func validarNumeroClienteTag(fl validator.FieldLevel) bool {
	numeroCliente := fl.Field().String()
	return ValidarNumeroCliente(numeroCliente)
}

func GetValidator() *validator.Validate {
	return validate
}
