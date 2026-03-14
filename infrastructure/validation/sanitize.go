package validation

import (
	"html"
	"regexp"
	"strings"
)

func SanitizeHTML(input string) string {
	input = html.EscapeString(input)

	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")

	return input
}

func SanitizeString(input string) string {
	input = strings.TrimSpace(input)

	// MED-03: Sanitización robusta — remover tags HTML y patrones peligrosos
	// Remover cualquier tag HTML
	reTag := regexp.MustCompile(`<[^>]*>`)
	input = reTag.ReplaceAllString(input, "")

	// Remover event handlers (case-insensitive, con variantes de encoding)
	reOnEvent := regexp.MustCompile(`(?i)\s*on\w+\s*=`)
	input = reOnEvent.ReplaceAllString(input, "")

	// Remover javascript: protocol (case-insensitive, con posibles null bytes)
	reJS := regexp.MustCompile(`(?i)j\s*a\s*v\s*a\s*s\s*c\s*r\s*i\s*p\s*t\s*:`)
	input = reJS.ReplaceAllString(input, "")

	// Remover operadores MongoDB peligrosos
	reMongo := regexp.MustCompile(`(?i)\$\w+`)
	input = reMongo.ReplaceAllString(input, "")

	return input
}

func SanitizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))

	re := regexp.MustCompile(`[^a-z0-9@._\-+]`)
	email = re.ReplaceAllString(email, "")

	return email
}

func SanitizeNumeroCliente(numero string) string {
	numero = strings.TrimSpace(numero)

	re := regexp.MustCompile(`[^a-zA-Z0-9\-]`)
	numero = re.ReplaceAllString(numero, "")

	return numero
}

func ValidarLongitud(input string, min, max int) bool {
	length := len(input)
	return length >= min && length <= max
}
