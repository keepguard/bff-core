package services

import (
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/domain/entities"
	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

// UserDomainService define as regras de negócio para usuários
type UserDomainService interface {
	ValidateUserCreation(user entities.User, company entities.Company) error
	GenerateCodeUser(user entities.User) string
	CanUserBeCreated(user entities.User) error
	ValidateUserType(userType valueobjects.UserType) error
	CanUserBeActivated(user entities.User) error
	CanUserBeBlocked(user entities.User) error
	ValidateUserUpdate(user entities.User, newName string, newEmail valueobjects.Email, newPhone valueobjects.Phone) error
}

type userDomainServiceImpl struct{}

// NewUserDomainService cria um novo serviço de domínio para usuários
func NewUserDomainService() UserDomainService {
	return &userDomainServiceImpl{}
}

// ValidateUserCreation valida se um usuário pode ser criado
func (s *userDomainServiceImpl) ValidateUserCreation(user entities.User, company entities.Company) error {
	// A empresa deve estar ativa para criar usuários
	if !company.IsActive() {
		return errors.New("company must be active to create users")
	}

	// A empresa não pode estar bloqueada
	if company.IsBlocked() {
		return errors.New("blocked company cannot create users")
	}

	// O usuário deve ter um tipo válido
	if err := s.ValidateUserType(user.UserType()); err != nil {
		return err
	}

	// O usuário deve ter um nome válido
	if user.Name() == "" {
		return errors.New("user name is required")
	}

	return nil
}

// GenerateCodeUser gera um código único para o usuário
func (s *userDomainServiceImpl) GenerateCodeUser(user entities.User) string {
	// Implementação simplificada - em produção usar algoritmo mais robusto
	userType := user.UserType().Value()
	companyID := user.CompanyID()

	// Formato: TIPO_COMPANYID_TIMESTAMP_RANDOM
	return userType + "_" + companyID[:8] + "_" + generateTimestamp() + "_" + generateRandomString(4)
}

// CanUserBeCreated verifica se um usuário pode ser criado
func (s *userDomainServiceImpl) CanUserBeCreated(user entities.User) error {
	// Verificar se o usuário tem dados válidos
	if user.Name() == "" {
		return errors.New("user name is required")
	}

	// Verificar se o email é válido
	if user.Email().Value() == "" {
		return errors.New("user email is required")
	}

	// Verificar se o telefone é válido
	if user.Phone().Value() == "" {
		return errors.New("user phone is required")
	}

	// Verificar se a empresa é válida
	if user.CompanyID() == "" {
		return errors.New("user company is required")
	}

	return nil
}

// ValidateUserType valida se o tipo de usuário é válido
func (s *userDomainServiceImpl) ValidateUserType(userType valueobjects.UserType) error {
	// Verificar se o tipo é válido
	if !userType.IsPerson() && !userType.IsCompany() {
		return errors.New("invalid user type")
	}

	return nil
}

// CanUserBeActivated verifica se um usuário pode ser ativado
func (s *userDomainServiceImpl) CanUserBeActivated(user entities.User) error {
	if user.IsActive() {
		return errors.New("user is already active")
	}

	if user.IsBlocked() {
		return errors.New("blocked user cannot be activated")
	}

	return nil
}

// CanUserBeBlocked verifica se um usuário pode ser bloqueado
func (s *userDomainServiceImpl) CanUserBeBlocked(user entities.User) error {
	if user.IsBlocked() {
		return errors.New("user is already blocked")
	}

	if !user.IsActive() {
		return errors.New("only active users can be blocked")
	}

	return nil
}

// ValidateUserUpdate valida se um usuário pode ser atualizado
func (s *userDomainServiceImpl) ValidateUserUpdate(user entities.User, newName string, newEmail valueobjects.Email, newPhone valueobjects.Phone) error {
	// Verificar se o usuário está ativo
	if !user.IsActive() {
		return errors.New("only active users can be updated")
	}

	// Verificar se o novo nome é válido
	if newName == "" {
		return errors.New("new name cannot be empty")
	}

	// Verificar se o novo email é válido
	if newEmail.Value() == "" {
		return errors.New("new email cannot be empty")
	}

	// Verificar se o novo telefone é válido
	if newPhone.Value() == "" {
		return errors.New("new phone cannot be empty")
	}

	return nil
}

// generateTimestamp gera um timestamp único
func generateTimestamp() string {
	// Implementação simplificada
	return "20240101120000" // Placeholder
}

// generateRandomString gera uma string aleatória
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
