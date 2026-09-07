package dto

import "time"

type UserConsentAcceptBatchRequestDTO struct {
	Email       string                    `json:"email"`
	AcceptedAt  time.Time                 `json:"acceptedAt"`
	Geolocation string                    `json:"geolocation,omitempty"`
	Consents    []UserConsentBatchItemDTO `json:"consents"`
}

type UserConsentBatchItemDTO struct {
	DocumentID  string `json:"documentId"`
	Version     int    `json:"version"`
	Accepted    bool   `json:"accepted"`
	ContentHash string `json:"contentHash,omitempty"`
}
