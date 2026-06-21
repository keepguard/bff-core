package enums

// MessageType representa os tipos de mensagem suportados
type MessageType string

const (
	// MessageTypeEmail representa mensagens via email
	MessageTypeEmail MessageType = "EMAIL"

	// MessageTypeSMS representa mensagens via SMS
	MessageTypeSMS MessageType = "SMS"

	// MessageTypePushNotification representa notificações push
	MessageTypePushNotification MessageType = "PUSH_NOTIFICATION"

	// MessageTypeWhatsApp representa mensagens via WhatsApp
	MessageTypeWhatsApp MessageType = "WHATSAPP"

	// MessageTypePush representa mensagens push
	MessageTypePush MessageType = "PUSH"
)

// String retorna a representação em string do MessageType
func (m MessageType) String() string {
	return string(m)
}

// IsValid verifica se o MessageType é válido
func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypeEmail, MessageTypeSMS, MessageTypePushNotification, MessageTypeWhatsApp, MessageTypePush:
		return true
	default:
		return false
	}
}
