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
	
	caracteresProhibidos := []string{
		"$where", "$regex", "$ne", "$gt", "$lt", "$gte", "$lte",
		"$in", "$nin", "$or", "$and", "$not", "$nor", "$exists",
		"javascript:", "onerror=", "onclick=", "onload=",
	}
	
	inputLower := strings.ToLower(input)
	for _, prohibido := range caracteresProhibidos {
		if strings.Contains(inputLower, prohibido) {
			input = strings.ReplaceAll(input, prohibido, "")
		}
	}
	
	reScript := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = reScript.ReplaceAllString(input, "")
	
	reIframe := regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`)
	input = reIframe.ReplaceAllString(input, "")
	
	reOnEvent := regexp.MustCompile(`(?i)\s*on\w+\s*=`)
	input = reOnEvent.ReplaceAllString(input, "")
	
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
