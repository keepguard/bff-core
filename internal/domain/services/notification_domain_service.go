package services

import (
	"errors"

	"github.com/keepguard/bff-core/internal/domain/entities"
)

// NotificationDomainService define as regras de negócio para notificações
type NotificationDomainService interface {
	ValidateNotificationPreferences(notification entities.UserNotification) error
	CanSendNotification(notification entities.UserNotification, channel string) error
	GetAvailableChannels(notification entities.UserNotification) []string
	ValidateChannel(channel string) error
	ShouldSendNotification(notification entities.UserNotification, channel string) error
}

type notificationDomainServiceImpl struct{}

// NewNotificationDomainService cria um novo serviço de domínio para notificações
func NewNotificationDomainService() NotificationDomainService {
	return &notificationDomainServiceImpl{}
}

// ValidateNotificationPreferences valida as preferências de notificação
func (s *notificationDomainServiceImpl) ValidateNotificationPreferences(notification entities.UserNotification) error {
	// Verificar se pelo menos um canal está habilitado
	if !notification.HasAnyNotificationEnabled() {
		return errors.New("at least one notification channel must be enabled")
	}

	// Verificar se o usuário é válido
	if notification.UserID() == "" {
		return errors.New("userID is required")
	}

	return nil
}

// CanSendNotification verifica se uma notificação pode ser enviada
func (s *notificationDomainServiceImpl) CanSendNotification(notification entities.UserNotification, channel string) error {
	// Verificar se o canal é válido
	if err := s.ValidateChannel(channel); err != nil {
		return err
	}

	// Verificar se o canal está habilitado para o usuário
	if !notification.IsChannelEnabled(channel) {
		return errors.New("channel is not enabled for this user")
	}

	return nil
}

// GetAvailableChannels retorna os canais disponíveis para o usuário
func (s *notificationDomainServiceImpl) GetAvailableChannels(notification entities.UserNotification) []string {
	return notification.GetEnabledChannels()
}

// ValidateChannel valida se o canal é válido
func (s *notificationDomainServiceImpl) ValidateChannel(channel string) error {
	validChannels := []string{"email", "sms", "push", "whatsapp"}

	for _, validChannel := range validChannels {
		if channel == validChannel {
			return nil
		}
	}

	return errors.New("invalid notification channel")
}

// ShouldSendNotification verifica se uma notificação deve ser enviada
func (s *notificationDomainServiceImpl) ShouldSendNotification(notification entities.UserNotification, channel string) error {
	// Verificar se o canal é válido
	if err := s.ValidateChannel(channel); err != nil {
		return err
	}

	// Verificar se o canal está habilitado
	if !notification.IsChannelEnabled(channel) {
		return errors.New("channel is not enabled")
	}

	// Verificar se pelo menos um canal está habilitado
	if !notification.HasAnyNotificationEnabled() {
		return errors.New("no notification channels are enabled")
	}

	return nil
}
