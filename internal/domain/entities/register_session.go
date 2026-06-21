package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-core/internal/domain/valueobjects"
)

// RegisterSession representa uma sessão de registro de usuário
type RegisterSession struct {
	id              string
	email           valueobjects.Email
	name            string
	password        string
	phone           valueobjects.Phone
	termsAccepted   bool
	termsVersion    string
	privacyAccepted bool
	privacyVersion  string
	userType        valueobjects.UserType
	status          RegisterSessionStatus
	token           string
	expiresAt       time.Time
	confirmedAt     *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// RegisterSessionStatus representa o status da sessão de registro
type RegisterSessionStatus string

const (
	RegisterSessionStatusPending   RegisterSessionStatus = "PENDING"
	RegisterSessionStatusConfirmed RegisterSessionStatus = "CONFIRMED"
	RegisterSessionStatusExpired   RegisterSessionStatus = "EXPIRED"
	RegisterSessionStatusCancelled RegisterSessionStatus = "CANCELLED"
)

// NewRegisterSession cria uma nova sessão de registro
func NewRegisterSession(
	email valueobjects.Email,
	name string,
	password string,
	phone valueobjects.Phone,
	termsAccepted bool,
	termsVersion string,
	privacyAccepted bool,
	privacyVersion string,
	userType valueobjects.UserType,
) (*RegisterSession, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	if !termsAccepted {
		return nil, errors.New("terms must be accepted")
	}

	if !privacyAccepted {
		return nil, errors.New("privacy must be accepted")
	}

	if termsVersion == "" {
		return nil, errors.New("termsVersion cannot be empty")
	}

	if privacyVersion == "" {
		return nil, errors.New("privacyVersion cannot be empty")
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // Expira em 24 horas

	return &RegisterSession{
		id:              generateRegisterSessionID(),
		email:           email,
		name:            name,
		password:        password,
		phone:           phone,
		termsAccepted:   termsAccepted,
		termsVersion:    termsVersion,
		privacyAccepted: privacyAccepted,
		privacyVersion:  privacyVersion,
		userType:        userType,
		status:          RegisterSessionStatusPending,
		token:           generateToken(),
		expiresAt:       expiresAt,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// Getters
func (rs *RegisterSession) ID() string {
	return rs.id
}

func (rs *RegisterSession) Email() valueobjects.Email {
	return rs.email
}

func (rs *RegisterSession) Name() string {
	return rs.name
}

func (rs *RegisterSession) Phone() valueobjects.Phone {
	return rs.phone
}

func (rs *RegisterSession) TermsAccepted() bool {
	return rs.termsAccepted
}

func (rs *RegisterSession) TermsVersion() string {
	return rs.termsVersion
}

func (rs *RegisterSession) PrivacyAccepted() bool {
	return rs.privacyAccepted
}

func (rs *RegisterSession) PrivacyVersion() string {
	return rs.privacyVersion
}

func (rs *RegisterSession) UserType() valueobjects.UserType {
	return rs.userType
}

func (rs *RegisterSession) Status() RegisterSessionStatus {
	return rs.status
}

func (rs *RegisterSession) Token() string {
	return rs.token
}

func (rs *RegisterSession) ExpiresAt() time.Time {
	return rs.expiresAt
}

func (rs *RegisterSession) ConfirmedAt() *time.Time {
	return rs.confirmedAt
}

func (rs *RegisterSession) CreatedAt() time.Time {
	return rs.createdAt
}

func (rs *RegisterSession) UpdatedAt() time.Time {
	return rs.updatedAt
}

// Business Methods
func (rs *RegisterSession) Confirm() error {
	if rs.status != RegisterSessionStatusPending {
		return errors.New("session is not pending")
	}

	if rs.IsExpired() {
		return errors.New("session has expired")
	}

	now := time.Now()
	rs.status = RegisterSessionStatusConfirmed
	rs.confirmedAt = &now
	rs.updatedAt = now
	return nil
}

func (rs *RegisterSession) Cancel() {
	rs.status = RegisterSessionStatusCancelled
	rs.updatedAt = time.Now()
}

func (rs *RegisterSession) MarkAsExpired() {
	rs.status = RegisterSessionStatusExpired
	rs.updatedAt = time.Now()
}

func (rs *RegisterSession) IsExpired() bool {
	return time.Now().After(rs.expiresAt)
}

func (rs *RegisterSession) IsPending() bool {
	return rs.status == RegisterSessionStatusPending
}

func (rs *RegisterSession) IsConfirmed() bool {
	return rs.status == RegisterSessionStatusConfirmed
}

func (rs *RegisterSession) IsCancelled() bool {
	return rs.status == RegisterSessionStatusCancelled
}

func (rs *RegisterSession) CanBeConfirmed() bool {
	return rs.status == RegisterSessionStatusPending && !rs.IsExpired()
}

func (rs *RegisterSession) GetRemainingTime() time.Duration {
	if rs.IsExpired() {
		return 0
	}
	return time.Until(rs.expiresAt)
}

func (rs *RegisterSession) IsValidToken(token string) bool {
	return rs.token == token && rs.status == RegisterSessionStatusPending && !rs.IsExpired()
}

// generateRegisterSessionID gera um ID único para a sessão de registro
func generateRegisterSessionID() string {
	// Implementação simplificada - em produção usar UUID
	return "reg_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// generateToken gera um token único para a sessão
func generateToken() string {
	// Implementação simplificada - em produção usar algoritmo mais robusto
	return "token_" + time.Now().Format("20060102150405") + "_" + randomString(12)
}
