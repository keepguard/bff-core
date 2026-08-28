package dto

import (
	"encoding/json"
	"time"
)

// UserConsentAcceptBatchRequestDTO é o contrato de POST /user-consents/accept-batch.
type UserConsentAcceptBatchRequestDTO struct {
	UserID      string                    `json:"userId"`
	Email       string                    `json:"email"`
	AcceptedAt  time.Time                 `json:"acceptedAt"`
	Geolocation string                    `json:"geolocation,omitempty"`
	Consents    []UserConsentBatchItemDTO `json:"consents"`
	ClientIP    string                    `json:"-"`
	UserAgent   string                    `json:"-"`
}

// UserConsentBatchItemDTO é um item do aceite seletivo.
type UserConsentBatchItemDTO struct {
	DocumentID  string `json:"documentId"`
	Version     int    `json:"version"`
	Accepted    bool   `json:"accepted"`
	ContentHash string `json:"contentHash,omitempty"`
}

// MarshalJSON formata acceptedAt no padrão do ms-user-consents (millis + Z).
func (dto UserConsentAcceptBatchRequestDTO) MarshalJSON() ([]byte, error) {
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
		UserID:      dto.UserID,
		Email:       dto.Email,
		AcceptedAt:  dto.AcceptedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Geolocation: dto.Geolocation,
		Consents:    make([]item, 0, len(dto.Consents)),
	}
	for _, c := range dto.Consents {
		payload.Consents = append(payload.Consents, item{
			DocumentID:  c.DocumentID,
			Version:     c.Version,
			Accepted:    c.Accepted,
			ContentHash: c.ContentHash,
		})
	}
	return json.Marshal(payload)
}
