package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMessageType_String testa a conversão para string
func TestMessageType_String(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"Email", MessageTypeEmail, "EMAIL"},
		{"SMS", MessageTypeSMS, "SMS"},
		{"Push Notification", MessageTypePushNotification, "PUSH_NOTIFICATION"},
		{"WhatsApp", MessageTypeWhatsApp, "WHATSAPP"},
		{"Push", MessageTypePush, "PUSH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msgType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMessageType_IsValid testa a validação dos tipos
func TestMessageType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected bool
	}{
		{"Email válido", MessageTypeEmail, true},
		{"SMS válido", MessageTypeSMS, true},
		{"Push Notification válido", MessageTypePushNotification, true},
		{"WhatsApp válido", MessageTypeWhatsApp, true},
		{"Push válido", MessageTypePush, true},
		{"Tipo inválido", MessageType("INVALID"), false},
		{"Tipo vazio", MessageType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msgType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
