package dto

import "time"

type ListPublishedConsentsQuery struct {
	TenantID      string
	CorrelationID string
}

type GetLatestConsentQuery struct {
	TenantID      string
	CorrelationID string
	ConsentType   string
}

type AcceptBatchConsentCommand struct {
	TenantID      string
	CorrelationID string
	Token         string
	UserID        string
	Email         string
	AcceptedAt    time.Time
	Geolocation   string
	Consents      []UserConsentBatchItemDTO
	ClientIP      string
	UserAgent     string
}

type ConsentDocumentViewDTO = ConsentDocumentResponseDTO
type UserConsentAcceptAllViewDTO = UserConsentAcceptAllResponseDTO
