package dto

import "time"

// UserConsentAcceptAllRequestDTO representa a requisição para aceitar todos os documentos de consentimento
type UserConsentAcceptAllRequestDTO struct {
	UserID      string    `json:"userId" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	AcceptedAt  time.Time `json:"acceptedAt" validate:"required"` // Formato: 2006-01-02T15:04:05.000Z
	Geolocation string    `json:"geolocation,omitempty"`
}

// MarshalJSON customiza a serialização JSON para incluir milissegundos
func (dto UserConsentAcceptAllRequestDTO) MarshalJSON() ([]byte, error) {
	type Alias UserConsentAcceptAllRequestDTO
	return []byte(`{` +
		`"userId":"` + dto.UserID + `",` +
		`"email":"` + dto.Email + `",` +
		`"acceptedAt":"` + dto.AcceptedAt.UTC().Format("2006-01-02T15:04:05.000Z") + `"` +
		(func() string {
			if dto.Geolocation != "" {
				return `,"geolocation":"` + dto.Geolocation + `"`
			}
			return ""
		})() +
		`}`), nil
}
