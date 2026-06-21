package dto

import (
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
)

// UserConsentAcceptAllResponseDTO representa a resposta do aceite em lote de consentimentos
type UserConsentAcceptAllResponseDTO struct {
	AcceptedConsents []AcceptedConsentItemDTO `json:"acceptedConsents"`
	TotalAccepted    int                      `json:"totalAccepted"`
}

// AcceptedConsentItemDTO representa um item de consentimento aceito
type AcceptedConsentItemDTO struct {
	ID                string                `json:"id"`
	UserID            string                `json:"userId"`
	Email             string                `json:"email"`
	ConsentDocumentID string                `json:"consentDocumentId"`
	Version           int                   `json:"version"`
	AcceptedAt        companyDto.CustomTime `json:"acceptedAt"`
	CreatedAt         companyDto.CustomTime `json:"createdAt"`
	IPAddress         string                `json:"ipAddress"`
	UserAgent         string                `json:"userAgent"`
	Geolocation       string                `json:"geolocation,omitempty"`
}
