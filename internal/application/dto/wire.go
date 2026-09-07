package dto

import (
	"encoding/json"
	"time"
)

// CustomTime faz parse de datas ISO com ou sem timezone (contrato dos ms-*).
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1]
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.99999",
		"2006-01-02T15:04:05.9999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
	}
	var err error
	for _, format := range formats {
		ct.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}
	return err
}

type MSUserCreateRequestDTO struct {
	CompanyID       string             `json:"companyId"`
	Type            string             `json:"type"`
	Email           string             `json:"email"`
	PhoneE164       string             `json:"phoneE164,omitempty"`
	PreferredLocale string             `json:"preferredLocale,omitempty"`
	Timezone        string             `json:"timezone,omitempty"`
	AvatarURL       string             `json:"avatarUrl,omitempty"`
	Status          string             `json:"status,omitempty"`
	PersonProfile   *PersonProfileDTO  `json:"personProfile,omitempty"`
	CompanyProfile  *CompanyProfileDTO `json:"companyProfile,omitempty"`
}

type PersonProfileDTO struct {
	FullName      string `json:"full_name"`
	DisplayHandle string `json:"display_handle,omitempty"`
	CPF           string `json:"cpf,omitempty"`
	RG            string `json:"rg,omitempty"`
	RGIssuer      string `json:"rg_issuer,omitempty"`
	RGState       string `json:"rg_state,omitempty"`
	DateOfBirth   string `json:"date_of_birth,omitempty"`
	Gender        string `json:"gender,omitempty"`
	MaritalStatus string `json:"marital_status,omitempty"`
	Nationality   string `json:"nationality,omitempty"`
	BirthCountry  string `json:"birth_country,omitempty"`
	BirthState    string `json:"birth_state,omitempty"`
	BirthCity     string `json:"birth_city,omitempty"`
	MotherName    string `json:"mother_name,omitempty"`
	FatherName    string `json:"father_name,omitempty"`
	PEP           bool   `json:"pep"`
	KYCStatus     string `json:"kyc_status,omitempty"`
	KYCLevel      string `json:"kyc_level,omitempty"`
	Occupation    string `json:"occupation,omitempty"`
	IncomeRange   string `json:"income_range,omitempty"`
}

type CompanyProfileDTO struct {
	CompanyID                     string `json:"company_id"`
	LegalNameSnapshot             string `json:"legal_name_snapshot,omitempty"`
	CNPJSnapshot                  string `json:"cnpj_snapshot,omitempty"`
	StateRegistrationSnapshot     string `json:"state_registration_snapshot,omitempty"`
	MunicipalRegistrationSnapshot string `json:"municipal_registration_snapshot,omitempty"`
	RepresentativeName            string `json:"representative_name,omitempty"`
	RepresentativeCPF             string `json:"representative_cpf,omitempty"`
}

type MSUserResponseDTO struct {
	ID              string            `json:"id"`
	CodeUser        string            `json:"codeUser"`
	CompanyID       string            `json:"companyId"`
	TenantId        string            `json:"tenantId"`
	Type            string            `json:"type"`
	Status          string            `json:"status"`
	Email           string            `json:"email"`
	PhoneE164       string            `json:"phoneE164,omitempty"`
	PreferredLocale string            `json:"preferredLocale,omitempty"`
	Timezone        string            `json:"timezone,omitempty"`
	AvatarURL       string            `json:"avatarUrl,omitempty"`
	DisplayHandle   string            `json:"displayHandle,omitempty"`
	PersonProfile   *PersonProfileDTO `json:"personProfile,omitempty"`
	CreatedAt       CustomTime        `json:"createdAt"`
	UpdatedAt       CustomTime        `json:"updatedAt"`
}

type MSUserRegisterInitRequestDTO struct {
	CompanyID                  string `json:"companyId"`
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

type MSUserRegisterConfirmRequestDTO struct {
	Email                 string `json:"email"`
	RegistrationSessionID string `json:"registrationSessionId"`
	Token                 string `json:"token"`
	EmailToken            string `json:"emailToken,omitempty"`
	SmsToken              string `json:"smsToken,omitempty"`
	WhatsAppToken         string `json:"whatsAppToken,omitempty"`
}

type MSUserRegisterResendRequestDTO struct {
	Email                 string `json:"email"`
	RegistrationSessionID string `json:"registrationSessionId"`
}

type MSUserRegisterInitResponseDTO struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	Email                 string `json:"email"`
	ExpiresIn             int    `json:"expiresIn"`
	Token                 string `json:"token"`
	EmailToken            string `json:"emailToken"`
	SmsToken              string `json:"smsToken"`
	WhatsAppToken         string `json:"whatsAppToken"`
}

type MSUserRegisterConfirmResponseDTO struct {
	RegistrationSessionID      string `json:"registration_session_id"`
	TenantId                   string `json:"tenant_id"`
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

type MSUserRegisterResendResponseDTO struct {
	Message                 string `json:"message"`
	Email                   string `json:"email"`
	NameFull                string `json:"nameFull"`
	Token                   string `json:"token"`
	RegistrationSessionID   string `json:"registrationSessionId"`
	ResendAttemptsRemaining int    `json:"resendAttemptsRemaining"`
	ExpiresIn               int    `json:"expiresIn"`
}

type MSUserNotifyCreateRequestDTO struct {
	UserID          string `json:"userId"`
	EmailEnabled    bool   `json:"emailEnabled"`
	SmsEnabled      bool   `json:"smsEnabled"`
	PushEnabled     bool   `json:"pushEnabled"`
	WhatsAppEnabled bool   `json:"whatsappEnabled"`
}

type MSUserNotifyResponseDTO struct {
	ID              string `json:"id"`
	UserID          string `json:"userId"`
	EmailEnabled    bool   `json:"emailEnabled"`
	SmsEnabled      bool   `json:"smsEnabled"`
	PushEnabled     bool   `json:"pushEnabled"`
	WhatsAppEnabled bool   `json:"whatsappEnabled"`
}

type CompanyMfaChannelDTO struct {
	ID       string `json:"id"`
	Channel  string `json:"channel"`
	Required bool   `json:"required"`
	Enabled  bool   `json:"enabled"`
}

type MSCompanyResponseDTO struct {
	ID          string                 `json:"id"`
	CodeCompany string                 `json:"codeCompany"`
	Name        string                 `json:"name"`
	LegalName   string                 `json:"legalName"`
	CNPJ        string                 `json:"cnpj"`
	Status      string                 `json:"status"`
	MfaChannels []CompanyMfaChannelDTO `json:"mfaChannels"`
	CreatedAt   CustomTime             `json:"createdAt"`
	UpdatedAt   CustomTime             `json:"updatedAt"`
	TenantId    string                 `json:"tenantId"`
}

type AuthUserCreateRequestDTO struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	IDUserExternal string `json:"id_user_external"`
	CodeUser       string `json:"code_user"`
	CompanyID      string `json:"company_id"`
	CompanyCode    string `json:"company_code"`
	TenantId       string `json:"tenant_id"`
}

type AuthUserCreateResponseDTO struct {
	ID            string      `json:"id"`
	CodeUser      string      `json:"codeUser"`
	Username      string      `json:"username"`
	Email         string      `json:"email"`
	FirstName     string      `json:"firstName"`
	LastName      string      `json:"lastName"`
	Phone         string      `json:"phone"`
	CompanyID     string      `json:"companyId"`
	CompanyCode   string      `json:"companyCode"`
	Type          string      `json:"type"`
	Status        string      `json:"status"`
	EmailVerified bool        `json:"emailVerified"`
	CreatedAt     CustomTime  `json:"createdAt"`
	UpdatedAt     CustomTime  `json:"updatedAt"`
	LastLoginAt   *CustomTime `json:"lastLoginAt,omitempty"`
	Roles         []string    `json:"roles"`
}

type AuthRegisterLoginRequestDTO struct {
	Username     string `json:"username" validate:"required"`
	PasswordHash string `json:"passwordHash" validate:"required"`
	CompanyID    string `json:"companyId,omitempty"`
	TenantId     string `json:"tenantId" validate:"required"`
}

type AuthLoginResponseDTO struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

type GenerateResetTokenMSRequestDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
}

type GenerateResetTokenMSResponseDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
	Token             string `json:"token"`
	ExpiresInSeconds  int64  `json:"expiresInSeconds"`
}

type UserByEmailResponseDTO struct {
	ID            string `json:"id"`
	CodeUser      string `json:"codeUser"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
}

type SendNotificationRequestDTO struct {
	UserID       string            `json:"userId"`
	TemplateType string            `json:"templateType"`
	Channel      string            `json:"channel"`
	Recipient    string            `json:"recipient"`
	Data         map[string]string `json:"data"`
}

type SendMessageRequestDTO struct {
	MessageType       string                 `json:"messageType"`
	CommunicationType string                 `json:"communicationType"`
	TemplateType      string                 `json:"templateType"`
	Recipient         string                 `json:"recipient"`
	Subject           string                 `json:"subject,omitempty"`
	Content           string                 `json:"content,omitempty"`
	CodeUser          string                 `json:"codeUser,omitempty"`
	Variables         map[string]interface{} `json:"variables"`
}

type SendMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ConsentDocumentResponseDTO struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Version       int    `json:"version"`
	Status        string `json:"status"`
	S3URL         string `json:"s3Url"`
	ContentHash   string `json:"contentHash"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
	CreatedBy     string `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
	UpdatedBy     string `json:"updatedBy,omitempty"`
	PublishedAt   string `json:"publishedAt,omitempty"`
}

type UserConsentAcceptRequestDTO struct {
	UserID            string `json:"userId"`
	Email             string `json:"email"`
	ConsentDocumentID string `json:"consentDocumentId"`
	Version           int    `json:"version"`
	AcceptedAt        string `json:"acceptedAt"`
	Geolocation       string `json:"geolocation,omitempty"`
}

type UserConsentResponseDTO struct {
	ID                string `json:"id"`
	UserID            string `json:"userId"`
	Email             string `json:"email"`
	ConsentDocumentID string `json:"consentDocumentId"`
	Version           int    `json:"version"`
	AcceptedAt        string `json:"acceptedAt"`
	IPAddress         string `json:"ipAddress"`
	UserAgent         string `json:"userAgent"`
	Geolocation       string `json:"geolocation,omitempty"`
	CreatedAt         string `json:"createdAt"`
}

type UserConsentAcceptAllRequestDTO struct {
	UserID      string    `json:"userId" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	AcceptedAt  time.Time `json:"acceptedAt" validate:"required"`
	Geolocation string    `json:"geolocation,omitempty"`
}

func (d UserConsentAcceptAllRequestDTO) MarshalJSON() ([]byte, error) {
	geo := ""
	if d.Geolocation != "" {
		geo = `,"geolocation":"` + d.Geolocation + `"`
	}
	return []byte(`{` +
		`"userId":"` + d.UserID + `",` +
		`"email":"` + d.Email + `",` +
		`"acceptedAt":"` + d.AcceptedAt.UTC().Format("2006-01-02T15:04:05.000Z") + `"` +
		geo +
		`}`), nil
}

type UserConsentAcceptAllResponseDTO struct {
	AcceptedConsents []AcceptedConsentItemDTO `json:"acceptedConsents"`
	TotalAccepted    int                      `json:"totalAccepted"`
}

type AcceptedConsentItemDTO struct {
	ID                string     `json:"id"`
	UserID            string     `json:"userId"`
	Email             string     `json:"email"`
	ConsentDocumentID string     `json:"consentDocumentId"`
	Version           int        `json:"version"`
	AcceptedAt        CustomTime `json:"acceptedAt"`
	CreatedAt         CustomTime `json:"createdAt"`
	IPAddress         string     `json:"ipAddress"`
	UserAgent         string     `json:"userAgent"`
	Geolocation       string     `json:"geolocation,omitempty"`
}

type UserConsentAcceptBatchRequestDTO struct {
	UserID      string                    `json:"userId"`
	Email       string                    `json:"email"`
	AcceptedAt  time.Time                 `json:"acceptedAt"`
	Geolocation string                    `json:"geolocation,omitempty"`
	Consents    []UserConsentBatchItemDTO `json:"consents"`
	ClientIP    string                    `json:"-"`
	UserAgent   string                    `json:"-"`
}

type UserConsentBatchItemDTO struct {
	DocumentID  string `json:"documentId"`
	Version     int    `json:"version"`
	Accepted    bool   `json:"accepted"`
	ContentHash string `json:"contentHash,omitempty"`
}

func (d UserConsentAcceptBatchRequestDTO) MarshalJSON() ([]byte, error) {
	type item struct {
		DocumentID  string `json:"documentId"`
		Version     int    `json:"version"`
		Accepted    bool   `json:"accepted"`
		ContentHash string `json:"contentHash,omitempty"`
	}
	payload := struct {
		UserID      string `json:"userId"`
		Email       string `json:"email"`
		AcceptedAt  string `json:"acceptedAt"`
		Geolocation string `json:"geolocation,omitempty"`
		Consents    []item `json:"consents"`
	}{
		UserID:      d.UserID,
		Email:       d.Email,
		AcceptedAt:  d.AcceptedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Geolocation: d.Geolocation,
		Consents:    make([]item, 0, len(d.Consents)),
	}
	for _, c := range d.Consents {
		payload.Consents = append(payload.Consents, item{
			DocumentID:  c.DocumentID,
			Version:     c.Version,
			Accepted:    c.Accepted,
			ContentHash: c.ContentHash,
		})
	}
	return json.Marshal(payload)
}
