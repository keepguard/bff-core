package dto

import (
	"context"
	"fmt"
)

// RegisterInitCommand comando para inicializar registro de usuário
type RegisterInitCommand struct {
	NameFull                   string
	Email                      string
	Password                   string
	ConfirmPassword            string
	Phone                      string
	HasAcceptedTermsAndPrivacy bool
	AcceptedMarketing          bool
	IPAddress                  string
	UserAgent                  string
	Geolocation                string
	Type                       string
	TenantId               string
	CorrelationID              string
	Context                    context.Context
}

// NewRegisterInitCommand cria um novo comando de inicialização de registro
func NewRegisterInitCommand(
	nameFull, email, password, confirmPassword, phone string,
	hasAcceptedTermsAndPrivacy, acceptedMarketing bool,
	ipAddress, userAgent, geolocation, userType, tenantId, correlationID string,
	ctx context.Context,
) RegisterInitCommand {
	return RegisterInitCommand{
		NameFull:                   nameFull,
		Email:                      email,
		Password:                   password,
		ConfirmPassword:            confirmPassword,
		Phone:                      phone,
		HasAcceptedTermsAndPrivacy: hasAcceptedTermsAndPrivacy,
		AcceptedMarketing:          acceptedMarketing,
		IPAddress:                  ipAddress,
		UserAgent:                  userAgent,
		Geolocation:                geolocation,
		Type:                       userType,
		TenantId:               tenantId,
		CorrelationID:              correlationID,
		Context:                    ctx,
	}
}

// Validate valida o comando
func (c RegisterInitCommand) Validate() error {
	if c.NameFull == "" {
		return fmt.Errorf("nameFull é obrigatório")
	}
	if c.Email == "" {
		return fmt.Errorf("email é obrigatório")
	}
	if c.Password == "" {
		return fmt.Errorf("password é obrigatório")
	}
	if len(c.Password) < 8 {
		return fmt.Errorf("password deve ter no mínimo 8 caracteres")
	}
	if c.ConfirmPassword == "" {
		return fmt.Errorf("confirmPassword é obrigatório")
	}
	if c.Password != c.ConfirmPassword {
		return fmt.Errorf("password e confirmPassword devem ser iguais")
	}
	if c.Phone == "" {
		return fmt.Errorf("phone é obrigatório")
	}
	if !c.HasAcceptedTermsAndPrivacy {
		return fmt.Errorf("aceitação dos termos e privacidade é obrigatória")
	}
	if c.Type == "" {
		return fmt.Errorf("type é obrigatório")
	}
	if c.Type != "PERSON" && c.Type != "COMPANY" {
		return fmt.Errorf("type deve ser PERSON ou COMPANY")
	}
	if c.TenantId == "" {
		return fmt.Errorf("tenantId é obrigatório")
	}
	if c.CorrelationID == "" {
		return fmt.Errorf("correlationId é obrigatório")
	}
	return nil
}

// RegisterConfirmCommand comando para confirmar registro de usuário
type RegisterConfirmCommand struct {
	Email                 string
	RegistrationSessionID string
	Token                 string
	TenantId          string
	CorrelationID         string
	Context               context.Context
}

// NewRegisterConfirmCommand cria um novo comando de confirmação de registro
func NewRegisterConfirmCommand(
	email, registrationSessionID, token, tenantId, correlationID string,
	ctx context.Context,
) RegisterConfirmCommand {
	return RegisterConfirmCommand{
		Email:                 email,
		RegistrationSessionID: registrationSessionID,
		Token:                 token,
		TenantId:          tenantId,
		CorrelationID:         correlationID,
		Context:               ctx,
	}
}

// Validate valida o comando
func (c RegisterConfirmCommand) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("email é obrigatório")
	}
	if c.RegistrationSessionID == "" {
		return fmt.Errorf("registrationSessionId é obrigatório")
	}
	if c.Token == "" {
		return fmt.Errorf("token é obrigatório")
	}
	if len(c.Token) != 6 {
		return fmt.Errorf("token deve ter exatamente 6 dígitos")
	}
	if c.TenantId == "" {
		return fmt.Errorf("tenantId é obrigatório")
	}
	if c.CorrelationID == "" {
		return fmt.Errorf("correlationId é obrigatório")
	}
	return nil
}

// RegisterResendCommand representa comando de reenvio de token
type RegisterResendCommand struct {
	Email                 string
	RegistrationSessionID string
	TenantId          string
	CorrelationID         string
	Context               context.Context
}

func NewRegisterResendCommand(email, registrationSessionID, tenantId, correlationID string, ctx context.Context) RegisterResendCommand {
	return RegisterResendCommand{
		Email:                 email,
		RegistrationSessionID: registrationSessionID,
		TenantId:          tenantId,
		CorrelationID:         correlationID,
		Context:               ctx,
	}
}

func (c RegisterResendCommand) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("email é obrigatório")
	}
	if c.RegistrationSessionID == "" {
		return fmt.Errorf("registrationSessionId é obrigatório")
	}
	return nil
}
