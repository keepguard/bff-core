package dto

// MSUserRegisterInitResponseDTO representa a resposta da inicialização de registro do ms-user
type MSUserRegisterInitResponseDTO struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	Email                 string `json:"email"`
	ExpiresIn             int    `json:"expiresIn"`
	Token                 string `json:"token"`
}

// MSUserRegisterConfirmResponseDTO representa a resposta da confirmação de registro do ms-user
type MSUserRegisterConfirmResponseDTO struct {
	RegistrationSessionID      string `json:"registration_session_id"`
	XApplication               string `json:"x_application"`
	Email                      string `json:"email"`
	NameFull                   string `json:"name_full"`
	Phone                      string `json:"phone"`
	Type                       string `json:"type"`
	HasAcceptedTermsAndPrivacy bool   `json:"has_accepted_terms_and_privacy"`
	AcceptedMarketing          bool   `json:"accepted_marketing"`
	IPAddress                  string `json:"ip_address"`
	UserAgent                  string `json:"user_agent"`
	Geolocation                string `json:"geolocation"`
	CreatedAt                  string `json:"created_at"`
	Attempts                   int    `json:"attempts"`
	Message                    string `json:"message"`
	PasswordHash               string `json:"passwordHash"`
}

// MSUserRegisterResendResponseDTO representa resposta de reenvio do ms-user
type MSUserRegisterResendResponseDTO struct {
	Message                 string `json:"message"`
	Email                   string `json:"email"`
	NameFull                string `json:"nameFull"`
	Token                   string `json:"token"`
	RegistrationSessionID   string `json:"registrationSessionId"`
	ResendAttemptsRemaining int    `json:"resendAttemptsRemaining"`
	ExpiresIn               int    `json:"expiresIn"`
}
