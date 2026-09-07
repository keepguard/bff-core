package consent

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
)

type ConsentPort interface {
	AcceptBatch(ctx context.Context, cmd appdto.AcceptBatchConsentCommand) (appdto.UserConsentAcceptAllViewDTO, error)
}

type service struct {
	consents port.UserConsentClient
}

func NewConsentPort(consents port.UserConsentClient) ConsentPort {
	if consents == nil {
		return nil
	}
	return &service{consents: consents}
}

func (s *service) AcceptBatch(ctx context.Context, cmd appdto.AcceptBatchConsentCommand) (appdto.UserConsentAcceptAllViewDTO, error) {
	req := appdto.UserConsentAcceptBatchRequestDTO{
		UserID:      cmd.UserID,
		Email:       cmd.Email,
		AcceptedAt:  cmd.AcceptedAt,
		Geolocation: cmd.Geolocation,
		Consents:    cmd.Consents,
		ClientIP:    cmd.ClientIP,
		UserAgent:   cmd.UserAgent,
	}
	return s.consents.AcceptBatch(ctx, req, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}
