package dto

// MSUserNotifyResponseDTO representa a resposta com preferências de notificação do ms-user
type MSUserNotifyResponseDTO struct {
	ID              string `json:"id"`
	UserID          string `json:"userId"`
	EmailEnabled    bool   `json:"emailEnabled"`
	SmsEnabled      bool   `json:"smsEnabled"`
	PushEnabled     bool   `json:"pushEnabled"`
	WhatsAppEnabled bool   `json:"whatsappEnabled"`
}
