package dto

// MSUserRegisterInitRequestDTO representa a requisição para inicializar registro no ms-user
type MSUserRegisterInitRequestDTO struct {
	Email                      string `json:"email"`
	NameFull                   string `json:"nameFull"`
	Password                   string `json:"password"`
	Phone                      string `json:"phone"`
	HasAcceptedTermsAndPrivacy bool   `json:"hasAcceptedTermsAndPrivacy"`
	AcceptedMarketing          bool   `json:"acceptedMarketing"`
	IPAddress                  string `json:"ipAddress"`
	UserAgent                  string `json:"userAgent"`
	Geolocation                string `json:"geolocation"`
	Type                       string `json:"type"`
}

// MSUserRegisterConfirmRequestDTO representa a requisição para confirmar registro no ms-user
type MSUserRegisterConfirmRequestDTO struct {
	Email                 string `json:"email"`
	RegistrationSessionID string `json:"registrationSessionId"`
	Token                 string `json:"token"`
	EmailToken            string `json:"emailToken,omitempty"`
	SmsToken              string `json:"smsToken,omitempty"`
	WhatsAppToken         string `json:"whatsAppToken,omitempty"`
}

// MSUserRegisterResendRequestDTO representa requisição de reenvio ao ms-user
type MSUserRegisterResendRequestDTO struct {
	Email                 string `json:"email"`
	RegistrationSessionID string `json:"registrationSessionId"`
}
