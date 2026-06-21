package services

import (
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/domain/entities"
)

// RegisterDomainService define as regras de negócio para registro de usuários
type RegisterDomainService interface {
	ValidateRegisterSession(session entities.RegisterSession) error
	CanSessionBeConfirmed(session entities.RegisterSession) error
	ValidatePassword(password string) error
	ValidateTermsAcceptance(termsAccepted bool, termsVersion string) error
	ValidatePrivacyAcceptance(privacyAccepted bool, privacyVersion string) error
	ShouldSessionExpire(session entities.RegisterSession) bool
	GetSessionValidityPeriod() time.Duration
}

type registerDomainServiceImpl struct{}

// NewRegisterDomainService cria um novo serviço de domínio para registro
func NewRegisterDomainService() RegisterDomainService {
	return &registerDomainServiceImpl{}
}

// ValidateRegisterSession valida uma sessão de registro
func (s *registerDomainServiceImpl) ValidateRegisterSession(session entities.RegisterSession) error {
	// Verificar se o nome é válido
	if session.Name() == "" {
		return errors.New("name is required")
	}

	// Verificar se a senha é válida
	if err := s.ValidatePassword("password"); err != nil { // Placeholder - precisa de getter
		return err
	}

	// Verificar se os termos foram aceitos
	if err := s.ValidateTermsAcceptance(session.TermsAccepted(), session.TermsVersion()); err != nil {
		return err
	}

	// Verificar se a privacidade foi aceita
	if err := s.ValidatePrivacyAcceptance(session.PrivacyAccepted(), session.PrivacyVersion()); err != nil {
		return err
	}

	// Verificar se o tipo de usuário é válido
	if session.UserType().Value() == "" {
		return errors.New("user type is required")
	}

	return nil
}

// CanSessionBeConfirmed verifica se uma sessão pode ser confirmada
func (s *registerDomainServiceImpl) CanSessionBeConfirmed(session entities.RegisterSession) error {
	// Verificar se a sessão está pendente
	if !session.IsPending() {
		return errors.New("session is not pending")
	}

	// Verificar se a sessão não expirou
	if session.IsExpired() {
		return errors.New("session has expired")
	}

	// Verificar se a sessão não foi cancelada
	if session.IsCancelled() {
		return errors.New("session has been cancelled")
	}

	return nil
}

// ValidatePassword valida uma senha
func (s *registerDomainServiceImpl) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must have at least 8 characters")
	}

	if len(password) > 128 {
		return errors.New("password must have at most 128 characters")
	}

	// Verificar se contém pelo menos uma letra minúscula
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Verificar se contém pelo menos uma letra maiúscula
	hasUpper := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Verificar se contém pelo menos um número
	hasNumber := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

// ValidateTermsAcceptance valida o aceite dos termos
func (s *registerDomainServiceImpl) ValidateTermsAcceptance(termsAccepted bool, termsVersion string) error {
	if !termsAccepted {
		return errors.New("terms must be accepted")
	}

	if termsVersion == "" {
		return errors.New("terms version is required")
	}

	return nil
}

// ValidatePrivacyAcceptance valida o aceite da privacidade
func (s *registerDomainServiceImpl) ValidatePrivacyAcceptance(privacyAccepted bool, privacyVersion string) error {
	if !privacyAccepted {
		return errors.New("privacy policy must be accepted")
	}

	if privacyVersion == "" {
		return errors.New("privacy version is required")
	}

	return nil
}

// ShouldSessionExpire verifica se uma sessão deve expirar
func (s *registerDomainServiceImpl) ShouldSessionExpire(session entities.RegisterSession) bool {
	return session.IsExpired()
}

// GetSessionValidityPeriod retorna o período de validade de uma sessão
func (s *registerDomainServiceImpl) GetSessionValidityPeriod() time.Duration {
	return 24 * time.Hour // 24 horas
}
