package dto

// MSUserNotifyCreateRequestDTO representa a requisição para criar preferências de notificação no ms-user
type MSUserNotifyCreateRequestDTO struct {
	UserID          string `json:"userId"`
	EmailEnabled    bool   `json:"emailEnabled"`
	SmsEnabled      bool   `json:"smsEnabled"`
	PushEnabled     bool   `json:"pushEnabled"`
	WhatsAppEnabled bool   `json:"whatsappEnabled"`
}
