package services

import (
	"errors"

	"github.com/keepguard/bff-core/internal/domain/entities"
	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

// CompanyDomainService define as regras de negócio para empresas
type CompanyDomainService interface {
	ValidateCompanyCreation(company entities.Company) error
	CanCompanyBeActivated(company entities.Company) error
	CanCompanyBeBlocked(company entities.Company) error
	ValidateXApplication(xApplication string) error
	CanCompanyCreateUsers(company entities.Company) error
	ValidateCompanyUpdate(company entities.Company, newName, newLegalName string, newCNPJ valueobjects.CNPJ) error
}

type companyDomainServiceImpl struct{}

// NewCompanyDomainService cria um novo serviço de domínio para empresas
func NewCompanyDomainService() CompanyDomainService {
	return &companyDomainServiceImpl{}
}

// ValidateCompanyCreation valida se uma empresa pode ser criada
func (s *companyDomainServiceImpl) ValidateCompanyCreation(company entities.Company) error {
	// Verificar se o X-Application é válido
	if err := s.ValidateXApplication(company.XApplication()); err != nil {
		return err
	}

	// Verificar se o nome é válido
	if company.Name() == "" {
		return errors.New("company name is required")
	}

	// Verificar se o nome legal é válido
	if company.LegalName() == "" {
		return errors.New("company legal name is required")
	}

	// Verificar se o CNPJ é válido
	if company.CNPJ().Value() == "" {
		return errors.New("company CNPJ is required")
	}

	return nil
}

// CanCompanyBeActivated verifica se uma empresa pode ser ativada
func (s *companyDomainServiceImpl) CanCompanyBeActivated(company entities.Company) error {
	if company.IsActive() {
		return errors.New("company is already active")
	}

	if company.IsBlocked() {
		return errors.New("blocked company cannot be activated")
	}

	return nil
}

// CanCompanyBeBlocked verifica se uma empresa pode ser bloqueada
func (s *companyDomainServiceImpl) CanCompanyBeBlocked(company entities.Company) error {
	if company.IsBlocked() {
		return errors.New("company is already blocked")
	}

	if !company.IsActive() {
		return errors.New("only active companies can be blocked")
	}

	return nil
}

// ValidateXApplication valida se o X-Application é válido
func (s *companyDomainServiceImpl) ValidateXApplication(xApplication string) error {
	if xApplication == "" {
		return errors.New("xApplication is required")
	}

	if len(xApplication) < 3 {
		return errors.New("xApplication must have at least 3 characters")
	}

	if len(xApplication) > 50 {
		return errors.New("xApplication must have at most 50 characters")
	}

	// Verificar se contém apenas caracteres válidos
	for _, char := range xApplication {
		if !isValidXApplicationChar(char) {
			return errors.New("xApplication contains invalid characters")
		}
	}

	return nil
}

// CanCompanyCreateUsers verifica se uma empresa pode criar usuários
func (s *companyDomainServiceImpl) CanCompanyCreateUsers(company entities.Company) error {
	if !company.IsActive() {
		return errors.New("company must be active to create users")
	}

	if company.IsBlocked() {
		return errors.New("blocked company cannot create users")
	}

	return nil
}

// ValidateCompanyUpdate valida se uma empresa pode ser atualizada
func (s *companyDomainServiceImpl) ValidateCompanyUpdate(company entities.Company, newName, newLegalName string, newCNPJ valueobjects.CNPJ) error {
	// Verificar se a empresa está ativa
	if !company.IsActive() {
		return errors.New("only active companies can be updated")
	}

	// Verificar se o novo nome é válido
	if newName == "" {
		return errors.New("new name cannot be empty")
	}

	// Verificar se o novo nome legal é válido
	if newLegalName == "" {
		return errors.New("new legal name cannot be empty")
	}

	// Verificar se o novo CNPJ é válido
	if newCNPJ.Value() == "" {
		return errors.New("new CNPJ cannot be empty")
	}

	return nil
}

// isValidXApplicationChar verifica se um caractere é válido para X-Application
func isValidXApplicationChar(char rune) bool {
	// Aceita letras, números, hífens e underscores
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' ||
		char == '_'
}
