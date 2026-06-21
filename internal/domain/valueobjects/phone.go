package valueobjects

import (
	"errors"
	"regexp"
)

// Phone representa um telefone válido
type Phone struct {
	value string
}

// NewPhone cria um novo Phone value object
func NewPhone(phone string) (Phone, error) {
	cleaned := cleanPhone(phone)
	if !isValidPhone(cleaned) {
		return Phone{}, errors.New("invalid phone format")
	}

	return Phone{value: cleaned}, nil
}

// Value retorna o valor do telefone
func (p Phone) Value() string {
	return p.value
}

// Formatted retorna o telefone formatado
func (p Phone) Formatted() string {
	// Formatar como (11) 99999-9999
	if len(p.value) == 11 {
		return "(" + p.value[:2] + ") " + p.value[2:7] + "-" + p.value[7:]
	} else if len(p.value) == 10 {
		return "(" + p.value[:2] + ") " + p.value[2:6] + "-" + p.value[6:]
	}
	return p.value
}

// String implementa a interface Stringer
func (p Phone) String() string {
	return p.value
}

// Equals verifica se dois telefones são iguais
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// AreaCode retorna o código de área
func (p Phone) AreaCode() string {
	if len(p.value) >= 2 {
		return p.value[:2]
	}
	return ""
}

// Number retorna o número sem o código de área
func (p Phone) Number() string {
	if len(p.value) >= 10 {
		return p.value[2:]
	}
	return p.value
}

// cleanPhone remove todos os caracteres não numéricos
func cleanPhone(phone string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(phone, "")
}

// isValidPhone valida o formato do telefone
func isValidPhone(phone string) bool {
	// Aceita 10 ou 11 dígitos (com DDD)
	return len(phone) == 10 || len(phone) == 11
}
