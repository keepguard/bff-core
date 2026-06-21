package dto

// UserConsentResponseDTO representa a resposta de consentimento de usuário
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
