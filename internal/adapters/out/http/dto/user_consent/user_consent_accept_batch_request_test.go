package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserConsentAcceptBatchRequestDTO_MarshalJSON(t *testing.T) {
	acceptedAt := time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC)
	dto := UserConsentAcceptBatchRequestDTO{
		UserID:      "11111111-1111-1111-1111-111111111111",
		Email:       "rafael@exemplo.com",
		AcceptedAt:  acceptedAt,
		Geolocation: "São Paulo, BR",
		ClientIP:    "189.45.12.8",
		UserAgent:   "KeepGuard/1",
		Consents: []UserConsentBatchItemDTO{
			{DocumentID: "doc-1", Version: 1, Accepted: true, ContentHash: "abc"},
		},
	}

	raw, err := json.Marshal(dto)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "2026-08-28T14:00:00.000Z", payload["acceptedAt"])
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", payload["userId"])
	assert.NotContains(t, payload, "clientIP")
	assert.NotContains(t, payload, "userAgent")
}
