package dto

// UserConsentAcceptRequestDTO representa a requisição para aceitar um consentimento
type UserConsentAcceptRequestDTO struct {
	UserID            string `json:"userId"`
	Email             string `json:"email"`
	ConsentDocumentID string `json:"consentDocumentId"`
	Version           int    `json:"version"`
	AcceptedAt        string `json:"acceptedAt"`
	Geolocation       string `json:"geolocation,omitempty"`
}
