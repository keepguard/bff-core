package enums

// CommunicationType representa os tipos de comunicação suportados
type CommunicationType string

const (
	// CommunicationTypeEmail representa comunicação via email
	CommunicationTypeEmail CommunicationType = "EMAIL"

	// CommunicationTypeSMS representa comunicação via SMS
	CommunicationTypeSMS CommunicationType = "SMS"

	// CommunicationTypePushNotification representa notificações push
	CommunicationTypePushNotification CommunicationType = "PUSH_NOTIFICATION"

	// CommunicationTypeWhatsApp representa comunicação via WhatsApp
	CommunicationTypeWhatsApp CommunicationType = "WHATSAPP"

	// CommunicationTypeTelegram representa comunicação via Telegram
	CommunicationTypeTelegram CommunicationType = "TELEGRAM"

	// CommunicationTypeSendGrid representa comunicação via SendGrid
	CommunicationTypeSendGrid CommunicationType = "SENDGRID"

	// CommunicationTypePush representa comunicação push
	CommunicationTypePush CommunicationType = "PUSH"
)

// String retorna a representação em string do CommunicationType
func (c CommunicationType) String() string {
	return string(c)
}

// IsValid verifica se o CommunicationType é válido
func (c CommunicationType) IsValid() bool {
	switch c {
	case CommunicationTypeEmail, CommunicationTypeSMS, CommunicationTypePushNotification,
		CommunicationTypeWhatsApp, CommunicationTypeTelegram, CommunicationTypeSendGrid, CommunicationTypePush:
		return true
	default:
		return false
	}
}
