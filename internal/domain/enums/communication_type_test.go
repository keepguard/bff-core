package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommunicationType_String testa a conversão para string
func TestCommunicationType_String(t *testing.T) {
	tests := []struct {
		name     string
		commType CommunicationType
		expected string
	}{
		{"Email", CommunicationTypeEmail, "EMAIL"},
		{"SMS", CommunicationTypeSMS, "SMS"},
		{"Push Notification", CommunicationTypePushNotification, "PUSH_NOTIFICATION"},
		{"WhatsApp", CommunicationTypeWhatsApp, "WHATSAPP"},
		{"Telegram", CommunicationTypeTelegram, "TELEGRAM"},
		{"SendGrid", CommunicationTypeSendGrid, "SENDGRID"},
		{"Push", CommunicationTypePush, "PUSH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.commType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCommunicationType_IsValid testa a validação dos tipos
func TestCommunicationType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		commType CommunicationType
		expected bool
	}{
		{"Email válido", CommunicationTypeEmail, true},
		{"SMS válido", CommunicationTypeSMS, true},
		{"Push Notification válido", CommunicationTypePushNotification, true},
		{"WhatsApp válido", CommunicationTypeWhatsApp, true},
		{"Telegram válido", CommunicationTypeTelegram, true},
		{"SendGrid válido", CommunicationTypeSendGrid, true},
		{"Push válido", CommunicationTypePush, true},
		{"Tipo inválido", CommunicationType("INVALID"), false},
		{"Tipo vazio", CommunicationType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.commType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
