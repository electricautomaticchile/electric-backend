package validation

import (
	"regexp"
	"unicode"
)

type PasswordError struct {
	Message string
}

func (e *PasswordError) Error() string {
	return e.Message
}

func ValidarPassword(password string) error {
	if len(password) < 8 {
		return &PasswordError{Message: "La contraseña debe tener al menos 8 caracteres"}
	}

	if len(password) > 128 {
		return &PasswordError{Message: "La contraseña no puede exceder 128 caracteres"}
	}

	tieneMayuscula := false
	tieneMinuscula := false
	tieneNumero := false
	tieneEspecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			tieneMayuscula = true
		case unicode.IsLower(char):
			tieneMinuscula = true
		case unicode.IsDigit(char):
			tieneNumero = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			tieneEspecial = true
		}
	}

	if !tieneMayuscula {
		return &PasswordError{Message: "La contraseña debe contener al menos una letra mayúscula"}
	}

	if !tieneMinuscula {
		return &PasswordError{Message: "La contraseña debe contener al menos una letra minúscula"}
	}

	if !tieneNumero {
		return &PasswordError{Message: "La contraseña debe contener al menos un número"}
	}

	if !tieneEspecial {
		return &PasswordError{Message: "La contraseña debe contener al menos un carácter especial"}
	}

	passwordsComunes := []string{
		"password", "Password1!", "12345678", "Qwerty123!", 
		"Admin123!", "Welcome1!", "Passw0rd!", "Chile123!",
	}

	for _, comun := range passwordsComunes {
		if password == comun {
			return &PasswordError{Message: "La contraseña es demasiado común"}
		}
	}

	patronesDebiles := []string{
		`^(.)\1{7,}$`,
		`^(012|123|234|345|456|567|678|789|890)+`,
		`^(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz)+`,
	}

	for _, patron := range patronesDebiles {
		matched, _ := regexp.MatchString(patron, password)
		if matched {
			return &PasswordError{Message: "La contraseña contiene patrones débiles"}
		}
	}

	return nil
}
