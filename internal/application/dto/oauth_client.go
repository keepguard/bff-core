package dto

type OAuthClientDTO struct {
	ID              string   `json:"id"`
	CompanyID       string   `json:"companyId"`
	ClientID        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret,omitempty"`
	Authorities     []string `json:"authorities"`
	Status          string   `json:"status"`
	TokenTTLSeconds int      `json:"tokenTtlSeconds"`
	Description     string   `json:"description,omitempty"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

type PaginatedOAuthClients struct {
	Content       []OAuthClientDTO `json:"content"`
	Page          int              `json:"page"`
	Size          int              `json:"size"`
	TotalElements int64            `json:"totalElements"`
	TotalPages    int              `json:"totalPages"`
	First         bool             `json:"first"`
	Last          bool             `json:"last"`
	HasNext       bool             `json:"hasNext"`
	HasPrevious   bool             `json:"hasPrevious"`
}

type OAuthClientCreateRequest struct {
	ClientID        string   `json:"clientId"`
	Description     string   `json:"description,omitempty"`
	Authorities     []string `json:"authorities,omitempty"`
	TokenTTLSeconds *int     `json:"tokenTtlSeconds,omitempty"`
}

type CollectorAgentRaw struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	CompanyID     string `json:"company_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CollectorType string `json:"collector_type"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type CollectorAgentDTO struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	CompanyID     string `json:"companyId"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CollectorType string `json:"collectorType"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type OAuthClientDetailResponse struct {
	OAuthClientDTO
	TenantID        string              `json:"tenantId,omitempty"`
	Agents          []CollectorAgentDTO `json:"agents"`
	AgentsLoadError string              `json:"agentsLoadError,omitempty"`
}
