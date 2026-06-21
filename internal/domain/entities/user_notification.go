package entities

import (
	"errors"
	"time"
)

// UserNotification representa as preferências de notificação de um usuário
type UserNotification struct {
	id              string
	userID          string
	emailEnabled    bool
	smsEnabled      bool
	pushEnabled     bool
	whatsAppEnabled bool
	createdAt       time.Time
	updatedAt       time.Time
}

// NewUserNotification cria uma nova preferência de notificação
func NewUserNotification(userID string, emailEnabled, smsEnabled, pushEnabled, whatsAppEnabled bool) (*UserNotification, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	now := time.Now()
	return &UserNotification{
		id:              generateNotificationID(),
		userID:          userID,
		emailEnabled:    emailEnabled,
		smsEnabled:      smsEnabled,
		pushEnabled:     pushEnabled,
		whatsAppEnabled: whatsAppEnabled,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// Getters
func (un *UserNotification) ID() string {
	return un.id
}

func (un *UserNotification) UserID() string {
	return un.userID
}

func (un *UserNotification) EmailEnabled() bool {
	return un.emailEnabled
}

func (un *UserNotification) SmsEnabled() bool {
	return un.smsEnabled
}

func (un *UserNotification) PushEnabled() bool {
	return un.pushEnabled
}

func (un *UserNotification) WhatsAppEnabled() bool {
	return un.whatsAppEnabled
}

func (un *UserNotification) CreatedAt() time.Time {
	return un.createdAt
}

func (un *UserNotification) UpdatedAt() time.Time {
	return un.updatedAt
}

// Business Methods
func (un *UserNotification) UpdatePreferences(emailEnabled, smsEnabled, pushEnabled, whatsAppEnabled bool) {
	un.emailEnabled = emailEnabled
	un.smsEnabled = smsEnabled
	un.pushEnabled = pushEnabled
	un.whatsAppEnabled = whatsAppEnabled
	un.updatedAt = time.Now()
}

func (un *UserNotification) EnableEmail() {
	un.emailEnabled = true
	un.updatedAt = time.Now()
}

func (un *UserNotification) DisableEmail() {
	un.emailEnabled = false
	un.updatedAt = time.Now()
}

func (un *UserNotification) EnableSMS() {
	un.smsEnabled = true
	un.updatedAt = time.Now()
}

func (un *UserNotification) DisableSMS() {
	un.smsEnabled = false
	un.updatedAt = time.Now()
}

func (un *UserNotification) EnablePush() {
	un.pushEnabled = true
	un.updatedAt = time.Now()
}

func (un *UserNotification) DisablePush() {
	un.pushEnabled = false
	un.updatedAt = time.Now()
}

func (un *UserNotification) EnableWhatsApp() {
	un.whatsAppEnabled = true
	un.updatedAt = time.Now()
}

func (un *UserNotification) DisableWhatsApp() {
	un.whatsAppEnabled = false
	un.updatedAt = time.Now()
}

func (un *UserNotification) HasAnyNotificationEnabled() bool {
	return un.emailEnabled || un.smsEnabled || un.pushEnabled || un.whatsAppEnabled
}

func (un *UserNotification) GetEnabledChannels() []string {
	channels := []string{}
	if un.emailEnabled {
		channels = append(channels, "email")
	}
	if un.smsEnabled {
		channels = append(channels, "sms")
	}
	if un.pushEnabled {
		channels = append(channels, "push")
	}
	if un.whatsAppEnabled {
		channels = append(channels, "whatsapp")
	}
	return channels
}

func (un *UserNotification) IsChannelEnabled(channel string) bool {
	switch channel {
	case "email":
		return un.emailEnabled
	case "sms":
		return un.smsEnabled
	case "push":
		return un.pushEnabled
	case "whatsapp":
		return un.whatsAppEnabled
	default:
		return false
	}
}

// generateNotificationID gera um ID único para a notificação
func generateNotificationID() string {
	// Implementação simplificada - em produção usar UUID
	return "notif_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}
