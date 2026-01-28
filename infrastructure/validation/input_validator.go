package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	sqlInjectionRegex = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|<script|onerror|onload)`)
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email no puede estar vacío")
	}
	if len(email) > 255 {
		return errors.New("email demasiado largo")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("formato de email inválido")
	}
	return nil
}

func ValidateStringLength(s string, minLen, maxLen int, fieldName string) error {
	s = strings.TrimSpace(s)
	if len(s) < minLen {
		return errors.New(fieldName + " es demasiado corto")
	}
	if len(s) > maxLen {
		return errors.New(fieldName + " es demasiado largo")
	}
	return nil
}

func ValidateAlphanumeric(s string, fieldName string) error {
	if !alphanumRegex.MatchString(s) {
		return errors.New(fieldName + " contiene caracteres no permitidos")
	}
	return nil
}

func ValidateNoSQLInjection(s string) error {
	if sqlInjectionRegex.MatchString(s) {
		return errors.New("entrada contiene patrones sospechosos")
	}
	return nil
}

func ValidateRequired(s string, fieldName string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New(fieldName + " es requerido")
	}
	return nil
}

func ValidateNumericString(s string, fieldName string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New(fieldName + " no puede estar vacío")
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return errors.New(fieldName + " debe contener solo números")
		}
	}
	return nil
}

func ValidateNombre(nombre string) error {
	if err := ValidateRequired(nombre, "nombre"); err != nil {
		return err
	}
	if err := ValidateStringLength(nombre, 2, 100, "nombre"); err != nil {
		return err
	}
	if err := ValidateNoSQLInjection(nombre); err != nil {
		return err
	}
	return nil
}

func ValidateDireccion(direccion string) error {
	if err := ValidateRequired(direccion, "dirección"); err != nil {
		return err
	}
	if err := ValidateStringLength(direccion, 5, 200, "dirección"); err != nil {
		return err
	}
	if err := ValidateNoSQLInjection(direccion); err != nil {
		return err
	}
	return nil
}

func ValidateComuna(comuna string) error {
	if err := ValidateRequired(comuna, "comuna"); err != nil {
		return err
	}
	if err := ValidateStringLength(comuna, 3, 50, "comuna"); err != nil {
		return err
	}
	return nil
}

func ValidateRegion(region string) error {
	if err := ValidateRequired(region, "región"); err != nil {
		return err
	}
	if err := ValidateStringLength(region, 3, 50, "región"); err != nil {
		return err
	}
	return nil
}
