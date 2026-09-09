package mapper

import (
	inboundDto "github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-core/internal/application/collector"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/knowledge"
	"github.com/keepguard/bff-core/internal/application/oauth"
)

func ToRegisterInitResponse(view appdto.RegisterInitViewDTO) inboundDto.RegisterInitResponseDTO {
	return inboundDto.RegisterInitResponseDTO{
		RegistrationSessionID: view.RegistrationSessionID,
		Email:                 view.Email,
		Phone:                 view.Phone,
		ExpiresIn:             view.ExpiresIn,
		RequiredChannels:      view.RequiredChannels,
		Token:                 view.Token,
		TokenExpiresIn:        view.TokenExpiresIn,
	}
}

func ToRegisterConfirmResponse(view appdto.RegisterConfirmViewDTO) inboundDto.RegisterConfirmResponseDTO {
	return inboundDto.RegisterConfirmResponseDTO{
		Token:          view.Token,
		TokenExpiresIn: view.TokenExpiresIn,
	}
}

func ToRegisterResendResponse(view appdto.RegisterResendViewDTO) inboundDto.RegisterResendResponseDTO {
	return inboundDto.RegisterResendResponseDTO{
		Message:                 view.Message,
		ResendAttemptsRemaining: view.ResendAttemptsRemaining,
		ExpiresIn:               view.ExpiresIn,
	}
}

func ToMeProfileResponse(view appdto.MeProfileViewDTO) inboundDto.MeProfileResponseDTO {
	resp := inboundDto.MeProfileResponseDTO{
		Email:           view.Email,
		PhoneE164:       view.PhoneE164,
		PreferredLocale: view.PreferredLocale,
		Timezone:        view.Timezone,
		AvatarURL:       view.AvatarURL,
		DisplayHandle:   view.DisplayHandle,
		Type:            view.Type,
		Status:          view.Status,
		CreatedAt:       view.CreatedAt,
	}
	if view.PersonProfile != nil {
		resp.PersonProfile = &inboundDto.MePersonProfileDTO{
			FullName: view.PersonProfile.FullName,
			HasCpf:   view.PersonProfile.HasCpf,
			CpfLast4: view.PersonProfile.CpfLast4,
		}
	}
	return resp
}

func ToAcceptBatchCommand(req inboundDto.UserConsentAcceptBatchRequestDTO, userID, token, tenantID, correlationID, clientIP, userAgent, geolocation string) appdto.AcceptBatchConsentCommand {
	consents := make([]appdto.UserConsentBatchItemDTO, 0, len(req.Consents))
	for _, item := range req.Consents {
		consents = append(consents, appdto.UserConsentBatchItemDTO{
			DocumentID:  item.DocumentID,
			Version:     item.Version,
			Accepted:    item.Accepted,
			ContentHash: item.ContentHash,
		})
	}
	return appdto.AcceptBatchConsentCommand{
		TenantID:      tenantID,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        userID,
		Email:         req.Email,
		AcceptedAt:    req.AcceptedAt,
		Geolocation:   geolocation,
		Consents:      consents,
		ClientIP:      clientIP,
		UserAgent:     userAgent,
	}
}

func ToGuardianExecuteActionRequest(req inboundDto.GuardianExecuteActionRequestDTO) appdto.GuardianExecuteActionRequest {
	return appdto.GuardianExecuteActionRequest{
		SuggestionID: req.SuggestionID,
		Confirmation: req.Confirmation,
	}
}

func ToGuardianRecipientUpsertRequest(req inboundDto.GuardianRecipientUpsertRequestDTO) appdto.GuardianRecipientUpsertRequest {
	return appdto.GuardianRecipientUpsertRequest{
		Email:   req.Email,
		Label:   req.Label,
		Enabled: req.Enabled,
	}
}

func ToCreateCollectorAgentCommand(req inboundDto.CollectorAgentCreateRequestDTO, companyFromCtx, tenantID, correlationID string) collector.CreateAgentCommand {
	return collector.CreateAgentCommand{
		CompanyFromCtx:  companyFromCtx,
		TenantID:        tenantID,
		CorrelationID:   correlationID,
		Name:            req.Name,
		Description:     req.Description,
		Context:         req.Context,
		CollectorType:   req.CollectorType,
		CollectorConfig: req.CollectorConfig,
		Prompt:          req.Prompt,
		Schedule:        req.Schedule,
		Enabled:         req.Enabled,
		DataSourceID:    req.DataSourceID,
	}
}

func ToUpdateCollectorAgentCommand(req inboundDto.CollectorAgentUpdateRequestDTO, companyFromCtx, tenantID, correlationID, id string) collector.UpdateAgentCommand {
	return collector.UpdateAgentCommand{
		CompanyFromCtx:  companyFromCtx,
		TenantID:        tenantID,
		CorrelationID:   correlationID,
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		Context:         req.Context,
		CollectorConfig: req.CollectorConfig,
		Prompt:          req.Prompt,
		Schedule:        req.Schedule,
		DataSourceID:    req.DataSourceID,
	}
}

func ToBulkCollectorAgentsCommand(req inboundDto.CollectorBulkRequestDTO, companyFromCtx, tenantID, correlationID string) collector.BulkAgentsCommand {
	return collector.BulkAgentsCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		Action:         req.Action,
		IDs:            req.IDs,
	}
}

func ToCreateCollectorDataSourceCommand(req inboundDto.CollectorDataSourceCreateRequestDTO, companyFromCtx, tenantID, correlationID string) collector.CreateDataSourceCommand {
	return collector.CreateDataSourceCommand{
		CompanyFromCtx:      companyFromCtx,
		TenantID:            tenantID,
		CorrelationID:       correlationID,
		Name:                req.Name,
		Slug:                req.Slug,
		Description:         req.Description,
		WebsiteURL:          req.WebsiteURL,
		CollectorType:       req.CollectorType,
		NameTemplate:        req.NameTemplate,
		DescriptionTemplate: req.DescriptionTemplate,
		PromptTemplate:      req.PromptTemplate,
		DefaultContext:      req.DefaultContext,
		DefaultSchedule:     req.DefaultSchedule,
		ConfigTemplate:      req.ConfigTemplate,
		Variables:           req.Variables,
		Notes:               req.Notes,
		Enabled:             req.Enabled,
		RateLimit:           req.RateLimit,
	}
}

func ToUpdateCollectorDataSourceCommand(req inboundDto.CollectorDataSourceUpdateRequestDTO, companyFromCtx, tenantID, correlationID, id string) collector.UpdateDataSourceCommand {
	return collector.UpdateDataSourceCommand{
		CompanyFromCtx:      companyFromCtx,
		TenantID:            tenantID,
		CorrelationID:       correlationID,
		ID:                  id,
		Name:                req.Name,
		Slug:                req.Slug,
		Description:         req.Description,
		WebsiteURL:          req.WebsiteURL,
		NameTemplate:        req.NameTemplate,
		DescriptionTemplate: req.DescriptionTemplate,
		PromptTemplate:      req.PromptTemplate,
		DefaultContext:      req.DefaultContext,
		DefaultSchedule:     req.DefaultSchedule,
		ConfigTemplate:      req.ConfigTemplate,
		Variables:           req.Variables,
		Notes:               req.Notes,
		RateLimit:           req.RateLimit,
	}
}

func ToPropagateCollectorDataSourceCommand(req inboundDto.CollectorPropagateRequestDTO, companyFromCtx, tenantID, correlationID, id string) collector.PropagateDataSourceCommand {
	return collector.PropagateDataSourceCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ID:             id,
		Fields:         req.Fields,
		DryRun:         req.DryRun,
		Limit:          req.Limit,
	}
}

func ToApplyCollectorIncidentSuccessorCommand(req inboundDto.CollectorApplySuccessorRequestDTO, companyFromCtx, tenantID, correlationID, actorUserID, id string) collector.ApplyIncidentSuccessorCommand {
	return collector.ApplyIncidentSuccessorCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		ActorUserID:    actorUserID,
		ID:             id,
		Confirmed:      req.Confirmed,
	}
}

func ToAskKnowledgeCommand(req inboundDto.KnowledgeAskRequestDTO, companyFromCtx, tenantID, correlationID string) knowledge.AskCommand {
	return knowledge.AskCommand{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		Question:       req.Question,
		Context:        req.Context,
	}
}

func oauthScope(companyFromCtx, tenantID, correlationID, token string) oauth.CompanyScope {
	return oauth.CompanyScope{
		CompanyFromCtx: companyFromCtx,
		TenantID:       tenantID,
		CorrelationID:  correlationID,
		BearerToken:    token,
	}
}

func ToOAuthCreateCommand(req inboundDto.OAuthClientCreateRequestDTO, companyFromCtx, tenantID, correlationID, token string) oauth.CreateClientCommand {
	return oauth.CreateClientCommand{
		CompanyScope: oauthScope(companyFromCtx, tenantID, correlationID, token),
		Body: appdto.OAuthClientCreateRequest{
			ClientID:        req.ClientID,
			Description:     req.Description,
			RoleID:          req.RoleID,
			TokenTTLSeconds: req.TokenTTLSeconds,
		},
	}
}

func ToOAuthUpdateCommand(req inboundDto.OAuthClientUpdateRequestDTO, companyFromCtx, tenantID, correlationID, token, id string) oauth.UpdateClientCommand {
	return oauth.UpdateClientCommand{
		CompanyScope: oauthScope(companyFromCtx, tenantID, correlationID, token),
		ID:           id,
		Body: appdto.OAuthClientUpdateRequest{
			Description:     req.Description,
			RoleID:          req.RoleID,
			TokenTTLSeconds: req.TokenTTLSeconds,
		},
	}
}
