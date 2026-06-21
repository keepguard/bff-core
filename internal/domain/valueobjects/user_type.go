package valueobjects

import (
	"errors"
	"strings"
)

// UserType representa um tipo de usuário válido
type UserType struct {
	value string
}

// Constantes para tipos de usuário (correspondem ao UserTypeEnum do ms-user)
const (
	UserTypePerson  = "PERSON"
	UserTypeCompany = "COMPANY"
)

// NewUserType cria um novo UserType value object
func NewUserType(userType string) (UserType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(userType))
	if !isValidUserType(normalized) {
		return UserType{}, errors.New("invalid user type")
	}

	return UserType{value: normalized}, nil
}

// Value retorna o valor do tipo de usuário
func (ut UserType) Value() string {
	return ut.value
}

// String implementa a interface Stringer
func (ut UserType) String() string {
	return ut.value
}

// Equals verifica se dois tipos de usuário são iguais
func (ut UserType) Equals(other UserType) bool {
	return ut.value == other.value
}

// IsPerson verifica se é uma pessoa física
func (ut UserType) IsPerson() bool {
	return ut.value == UserTypePerson
}

// IsCompany verifica se é uma pessoa jurídica
func (ut UserType) IsCompany() bool {
	return ut.value == UserTypeCompany
}

// HasPermission verifica se o tipo de usuário tem uma permissão específica
func (ut UserType) HasPermission(permission string) bool {
	switch ut.value {
	case UserTypePerson:
		return permission == "READ" || permission == "WRITE"
	case UserTypeCompany:
		return permission == "READ" || permission == "WRITE"
	default:
		return false
	}
}

// isValidUserType valida se o tipo de usuário é válido
func isValidUserType(userType string) bool {
	validTypes := []string{UserTypePerson, UserTypeCompany}
	for _, validType := range validTypes {
		if userType == validType {
			return true
		}
	}
	return false
}
