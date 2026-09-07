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
