package dto

// RegisterInitRequestDTO representa a requisição para inicializar o registro de usuário
type RegisterInitRequestDTO struct {
	Email                      string `json:"email" validate:"required,email"`
	NameFull                   string `json:"nameFull" validate:"required,min=3,max=255"`
	Password                   string `json:"password" validate:"required,min=8,max=100"`
	ConfirmPassword            string `json:"confirmPassword" validate:"required,min=8,max=100,eqfield=Password"`
	Phone                      string `json:"phone" validate:"required,min=10,max=20"`
	HasAcceptedTermsAndPrivacy bool   `json:"hasAcceptedTermsAndPrivacy" validate:"required"`
	AcceptedMarketing          *bool  `json:"acceptedMarketing,omitempty"`
	IPAddress                  string `json:"ipAddress,omitempty" validate:"omitempty,max=45"`
	UserAgent                  string `json:"userAgent,omitempty" validate:"omitempty,max=500"`
	Geolocation                string `json:"geolocation,omitempty" validate:"omitempty,max=255"`
	Type                       string `json:"type" validate:"required,oneof=PERSON COMPANY"`
}

// RegisterInitResponseDTO representa a resposta da inicialização do registro
type RegisterInitResponseDTO struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	Email                 string `json:"email"`
	ExpiresIn             int    `json:"expiresIn"`
	Token                 string `json:"token,omitempty"`
	TokenExpiresIn        int64  `json:"tokenExpiresIn,omitempty"`
}

// RegisterConfirmRequestDTO representa a requisição para confirmar o registro de usuário
type RegisterConfirmRequestDTO struct {
	Email                 string `json:"email" validate:"required,email"`
	RegistrationSessionID string `json:"registrationSessionId" validate:"required,uuid"`
	Token                 string `json:"token" validate:"required,len=6"`
}

// RegisterConfirmResponseDTO representa a resposta da confirmação do registro
type RegisterConfirmResponseDTO struct {
	Token          string `json:"token"`
	TokenExpiresIn int64  `json:"tokenExpiresIn"`
}

// RegisterResendRequestDTO representa requisição de reenvio de token
type RegisterResendRequestDTO struct {
	Email                 string `json:"email" validate:"required,email"`
	RegistrationSessionID string `json:"registrationSessionId" validate:"required,uuid"`
}

// RegisterResendResponseDTO representa resposta de reenvio de token
type RegisterResendResponseDTO struct {
	Message                 string `json:"message"`
	ResendAttemptsRemaining int    `json:"resendAttemptsRemaining"`
	ExpiresIn               int    `json:"expiresIn"`
}
