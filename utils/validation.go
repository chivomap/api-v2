package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidateQuery valida y sanitiza query parameters
func ValidateQuery(query string) (string, bool) {
	if query == "" {
		return "", false
	}

	// Remover espacios extras
	query = strings.TrimSpace(query)
	
	// Verificar longitud
	if len(query) > 100 {
		return "", false
	}

	// Verificar que no contenga caracteres peligrosos
	if containsDangerousChars(query) {
		return "", false
	}

	// Verificar que sea UTF-8 válido
	if !utf8.ValidString(query) {
		return "", false
	}

	return query, true
}

// ValidateWhatIs valida el parámetro whatIs
func ValidateWhatIs(whatIs string) (string, bool) {
	if whatIs == "" {
		return "", false
	}

	whatIs = strings.TrimSpace(strings.ToUpper(whatIs))
	
	// Lista de códigos permitidos según la documentación original
	allowedValues := map[string]bool{
		"D":   true, // Departamentos
		"M":   true, // Municipios  
		"NAM": true, // Nombres/ubicaciones (distritos, cantones, caseríos)
	}

	if !allowedValues[whatIs] {
		return "", false
	}

	return whatIs, true
}

// containsDangerousChars verifica caracteres peligrosos para SQL injection
func containsDangerousChars(input string) bool {
	// Patrones peligrosos comunes
	dangerousPatterns := []string{
		`'`,           // Single quote
		`"`,           // Double quote
		`;`,           // Semicolon
		`--`,          // SQL comment
		`/*`,          // SQL comment start
		`*/`,          // SQL comment end
		`\x00`,        // Null byte
		`\x1a`,        // Substitute character
		`<script`,     // XSS básico
		`javascript:`, // XSS
		`<iframe`,     // XSS
		`onclick`,     // XSS
		`onload`,      // XSS
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}

	// Verificar patrones con regex
	sqlInjectionPattern := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`)
	if sqlInjectionPattern.MatchString(input) {
		return true
	}

	return false
}

// SanitizeString limpia y sanitiza strings de entrada
func SanitizeString(input string) string {
	// Remover espacios extra
	input = strings.TrimSpace(input)
	
	// Reemplazar múltiples espacios con uno solo
	spaceRegex := regexp.MustCompile(`\s+`)
	input = spaceRegex.ReplaceAllString(input, " ")
	
	return input
}