package dto

import "encoding/json"

type OAuthClientDTO struct {
	ID              string   `json:"id"`
	CompanyID       string   `json:"companyId"`
	ClientID        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret,omitempty"`
	ServiceRoleID   string   `json:"serviceRoleId,omitempty"`
	ServiceRoleName string   `json:"serviceRoleName,omitempty"`
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
	ClientID        string `json:"clientId"`
	Description     string `json:"description,omitempty"`
	RoleID          string `json:"roleId"`
	TokenTTLSeconds *int   `json:"tokenTtlSeconds,omitempty"`
}

type OAuthClientUpdateRequest struct {
	Description     string `json:"description,omitempty"`
	RoleID          string `json:"roleId"`
	TokenTTLSeconds *int   `json:"tokenTtlSeconds,omitempty"`
}

type OAuthServiceRoleAuthorityDTO struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type OAuthServiceRoleDTO struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Authorities []OAuthServiceRoleAuthorityDTO `json:"authorities"`
}

type CollectorAgentRaw struct {
	ID              string                     `json:"id"`
	Code            string                     `json:"code"`
	CompanyID       string                     `json:"company_id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	Context         string                     `json:"context"`
	CollectorType   string                     `json:"collector_type"`
	CollectorConfig json.RawMessage            `json:"collector_config,omitempty"`
	Prompt          string                     `json:"prompt,omitempty"`
	Schedule        json.RawMessage            `json:"schedule,omitempty"`
	Enabled         bool                       `json:"enabled"`
	DataSourceID    string                     `json:"data_source_id,omitempty"`
	DataSourceSlug  string                     `json:"data_source_slug,omitempty"`
	DataSourceName  string                     `json:"data_source_name,omitempty"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	LastExecution   *CollectorLastExecutionRaw `json:"last_execution,omitempty"`
	OpenIncident    *CollectorOpenIncidentRaw  `json:"open_incident,omitempty"`
}

type CollectorOpenIncidentRaw struct {
	IncidentID     string `json:"incident_id"`
	Classification string `json:"classification"`
	Occurrences    int    `json:"occurrences"`
}

type CollectorOpenIncidentDTO struct {
	IncidentID     string `json:"incidentId"`
	Classification string `json:"classification"`
	Occurrences    int    `json:"occurrences"`
}

type CollectorLastExecutionRaw struct {
	ID         string `json:"id"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	Status     string `json:"status"`
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
	TenantID string `json:"tenantId,omitempty"`
}
