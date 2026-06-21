package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Email representa um email válido
type Email struct {
	value string
}

// NewEmail cria um novo Email value object
func NewEmail(email string) (Email, error) {
	if !isValidEmail(email) {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: strings.ToLower(strings.TrimSpace(email))}, nil
}

// Value retorna o valor do email
func (e Email) Value() string {
	return e.value
}

// Domain retorna o domínio do email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart retorna a parte local do email (antes do @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// String implementa a interface Stringer
func (e Email) String() string {
	return e.value
}

// Equals verifica se dois emails são iguais
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// isValidEmail valida o formato do email
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Padrão básico de validação de email
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}
