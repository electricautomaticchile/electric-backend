package validation

import (
	"regexp"
	"strings"
)

func ValidarTelefonoChileno(telefono string) bool {
	if telefono == "" {
		return false
	}

	telefono = strings.TrimSpace(telefono)
	telefono = strings.ReplaceAll(telefono, " ", "")
	telefono = strings.ReplaceAll(telefono, "-", "")
	telefono = strings.ReplaceAll(telefono, "(", "")
	telefono = strings.ReplaceAll(telefono, ")", "")

	patronesValidos := []string{
		`^\+569\d{8}$`,
		`^569\d{8}$`,
		`^9\d{8}$`,
		`^\+562\d{8}$`,
		`^2\d{8}$`,
	}

	for _, patron := range patronesValidos {
		matched, _ := regexp.MatchString(patron, telefono)
		if matched {
			return true
		}
	}

	return false
}

func NormalizarTelefono(telefono string) string {
	if telefono == "" {
		return ""
	}

	telefono = strings.TrimSpace(telefono)
	telefono = strings.ReplaceAll(telefono, " ", "")
	telefono = strings.ReplaceAll(telefono, "-", "")
	telefono = strings.ReplaceAll(telefono, "(", "")
	telefono = strings.ReplaceAll(telefono, ")", "")

	if strings.HasPrefix(telefono, "+56") {
		return telefono
	}

	if strings.HasPrefix(telefono, "56") && len(telefono) >= 11 {
		return "+" + telefono
	}

	if strings.HasPrefix(telefono, "9") && len(telefono) == 9 {
		return "+56" + telefono
	}

	if strings.HasPrefix(telefono, "2") && len(telefono) == 9 {
		return "+56" + telefono
	}

	return telefono
}
