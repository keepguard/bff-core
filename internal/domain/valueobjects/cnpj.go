package valueobjects

import (
	"errors"
	"regexp"
	"strconv"
)

// CNPJ representa um CNPJ válido
type CNPJ struct {
	value string
}

// NewCNPJ cria um novo CNPJ value object
func NewCNPJ(cnpj string) (CNPJ, error) {
	cleaned := cleanCNPJ(cnpj)
	if !isValidCNPJ(cleaned) {
		return CNPJ{}, errors.New("invalid CNPJ format")
	}

	return CNPJ{value: cleaned}, nil
}

// Value retorna o valor do CNPJ
func (c CNPJ) Value() string {
	return c.value
}

// Formatted retorna o CNPJ formatado
func (c CNPJ) Formatted() string {
	// Formatar como 12.345.678/0001-90
	if len(c.value) == 14 {
		return c.value[:2] + "." + c.value[2:5] + "." + c.value[5:8] + "/" + c.value[8:12] + "-" + c.value[12:]
	}
	return c.value
}

// String implementa a interface Stringer
func (c CNPJ) String() string {
	return c.value
}

// Equals verifica se dois CNPJs são iguais
func (c CNPJ) Equals(other CNPJ) bool {
	return c.value == other.value
}

// cleanCNPJ remove todos os caracteres não numéricos
func cleanCNPJ(cnpj string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(cnpj, "")
}

// isValidCNPJ valida o formato do CNPJ
func isValidCNPJ(cnpj string) bool {
	if len(cnpj) != 14 {
		return false
	}

	// Verificar se todos os dígitos são iguais
	if allDigitsEqual(cnpj) {
		return false
	}

	// Calcular dígitos verificadores
	firstDigit := calculateCNPJFirstDigit(cnpj[:12])
	secondDigit := calculateCNPJSecondDigit(cnpj[:13])

	expected := strconv.Itoa(firstDigit) + strconv.Itoa(secondDigit)
	return cnpj[12:] == expected
}

// allDigitsEqual verifica se todos os dígitos são iguais
func allDigitsEqual(cnpj string) bool {
	first := cnpj[0]
	for _, digit := range cnpj {
		if rune(first) != digit {
			return false
		}
	}
	return true
}

// calculateCNPJFirstDigit calcula o primeiro dígito verificador do CNPJ
func calculateCNPJFirstDigit(cnpj string) int {
	weights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i, digit := range cnpj {
		digitValue, _ := strconv.Atoi(string(digit))
		sum += digitValue * weights[i]
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}

// calculateCNPJSecondDigit calcula o segundo dígito verificador do CNPJ
func calculateCNPJSecondDigit(cnpj string) int {
	weights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i, digit := range cnpj {
		digitValue, _ := strconv.Atoi(string(digit))
		sum += digitValue * weights[i]
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
