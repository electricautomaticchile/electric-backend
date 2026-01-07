package validation

import (
	"regexp"
	"strconv"
	"strings"
)

func ValidarRUT(rut string) bool {
	if rut == "" {
		return false
	}

	rut = strings.ReplaceAll(rut, ".", "")
	rut = strings.ReplaceAll(rut, "-", "")
	rut = strings.ToUpper(strings.TrimSpace(rut))

	if len(rut) < 2 {
		return false
	}

	matched, _ := regexp.MatchString(`^[0-9]{7,8}[0-9K]$`, rut)
	if !matched {
		return false
	}

	dv := rut[len(rut)-1:]
	numero := rut[:len(rut)-1]

	dvCalculado := calcularDVRUT(numero)
	return dv == dvCalculado
}

func calcularDVRUT(rut string) string {
	suma := 0
	multiplicador := 2

	for i := len(rut) - 1; i >= 0; i-- {
		digito, _ := strconv.Atoi(string(rut[i]))
		suma += digito * multiplicador
		multiplicador++
		if multiplicador > 7 {
			multiplicador = 2
		}
	}

	resto := suma % 11
	dv := 11 - resto

	if dv == 11 {
		return "0"
	}
	if dv == 10 {
		return "K"
	}
	return strconv.Itoa(dv)
}

func NormalizarRUT(rut string) string {
	rut = strings.ReplaceAll(rut, ".", "")
	rut = strings.ReplaceAll(rut, "-", "")
	rut = strings.ToUpper(strings.TrimSpace(rut))
	
	if len(rut) < 2 {
		return rut
	}
	
	dv := rut[len(rut)-1:]
	numero := rut[:len(rut)-1]
	
	return numero + "-" + dv
}
