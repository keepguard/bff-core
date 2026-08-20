package dto

import (
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
)

// MSUserResponseDTO representa a resposta com dados do usuário do ms-user
type MSUserResponseDTO struct {
	ID              string                `json:"id"`
	CodeUser        string                `json:"codeUser"`
	CompanyID       string                `json:"companyId"`
	TenantId    string                `json:"tenantId"`
	Type            string                `json:"type"`
	Status          string                `json:"status"`
	Email           string                `json:"email"`
	PhoneE164       string                `json:"phoneE164,omitempty"`
	PreferredLocale string                `json:"preferredLocale,omitempty"`
	Timezone        string                `json:"timezone,omitempty"`
	AvatarURL       string                `json:"avatarUrl,omitempty"`
	CreatedAt       companyDto.CustomTime `json:"createdAt"`
	UpdatedAt       companyDto.CustomTime `json:"updatedAt"`
}
