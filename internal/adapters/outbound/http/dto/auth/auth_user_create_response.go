package dto

import (
	companyDto "github.com/keepguard/bff-core/internal/adapters/outbound/http/dto/company"
)

// AuthUserCreateResponseDTO representa a resposta da criação de usuário no ms-auth
type AuthUserCreateResponseDTO struct {
	ID            string                 `json:"id"`
	CodeUser      string                 `json:"codeUser"`
	Username      string                 `json:"username"`
	Email         string                 `json:"email"`
	FirstName     string                 `json:"firstName"`
	LastName      string                 `json:"lastName"`
	Phone         string                 `json:"phone"`
	CompanyID     string                 `json:"companyId"`
	CompanyCode   string                 `json:"companyCode"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	EmailVerified bool                   `json:"emailVerified"`
	CreatedAt     companyDto.CustomTime  `json:"createdAt"`
	UpdatedAt     companyDto.CustomTime  `json:"updatedAt"`
	LastLoginAt   *companyDto.CustomTime `json:"lastLoginAt,omitempty"`
	Roles         []string               `json:"roles"`
}
