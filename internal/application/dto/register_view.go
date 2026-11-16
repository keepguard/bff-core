package dto

// RegisterInitViewDTO é a visão de aplicação da inicialização de registro.
type RegisterInitViewDTO struct {
	RegistrationSessionID string
	Email                 string
	Phone                 string
	ExpiresIn             int
	RequiredChannels      []string
	Token                 string
	TokenExpiresIn        int64
	// ResendAvailable é preenchido só com saga_hardening ligado (aditivo): true indica que o reenvio está disponível.
	ResendAvailable *bool
}

// RegisterConfirmViewDTO é a visão de aplicação da confirmação de registro.
type RegisterConfirmViewDTO struct {
	Token          string
	TokenExpiresIn int64
}

// RegisterResendViewDTO é a visão de aplicação do reenvio de token.
type RegisterResendViewDTO struct {
	Message                 string
	ResendAttemptsRemaining int
	ExpiresIn               int
}
