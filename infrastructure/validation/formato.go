package validation

import (
	"regexp"
	"strings"
)

func ValidarPeriodo(periodo string) bool {
	if periodo == "" {
		return false
	}

	periodo = strings.TrimSpace(periodo)

	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, periodo)
	if !matched {
		return false
	}

	if len(periodo) != 7 {
		return false
	}

	mes := periodo[5:7]
	mesInt := 0
	if len(mes) == 2 {
		if mes[0] >= '0' && mes[0] <= '9' && mes[1] >= '0' && mes[1] <= '9' {
			mesInt = int(mes[0]-'0')*10 + int(mes[1]-'0')
		}
	}

	if mesInt < 1 || mesInt > 12 {
		return false
	}

	return true
}

func ValidarNumeroCliente(numeroCliente string) bool {
	if numeroCliente == "" {
		return false
	}

	numeroCliente = strings.TrimSpace(numeroCliente)

	if len(numeroCliente) < 5 || len(numeroCliente) > 20 {
		return false
	}

	matched, _ := regexp.MatchString(`^[A-Z0-9\-]+$`, strings.ToUpper(numeroCliente))
	return matched
}
