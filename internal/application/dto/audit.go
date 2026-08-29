package dto

import "time"

type AuditActor struct {
	Type     string   `json:"type"`
	CodeUser string   `json:"codeUser,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	ClientIP string   `json:"clientIp,omitempty"`
	DeviceID string   `json:"deviceId,omitempty"`
}

type AuditResource struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type AuditChange struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type AuditEventResponse struct {
	EventID       string         `json:"eventId"`
	OccurredAt    time.Time      `json:"occurredAt"`
	SchemaVersion int            `json:"schemaVersion"`
	SourceService string         `json:"sourceService"`
	CorrelationID string         `json:"correlationId"`
	RequestID     string         `json:"requestId,omitempty"`
	TenantID      string         `json:"tenantId,omitempty"`
	CompanyID     string         `json:"companyId,omitempty"`
	Actor         AuditActor     `json:"actor"`
	Action        string         `json:"action"`
	Resource      AuditResource  `json:"resource"`
	Outcome       string         `json:"outcome"`
	Reason        string         `json:"reason,omitempty"`
	Changes       []AuditChange  `json:"changes,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type JourneyHopResponse struct {
	EventID       string    `json:"eventId"`
	OccurredAt    time.Time `json:"occurredAt"`
	SourceService string    `json:"sourceService"`
	Action        string    `json:"action"`
	Outcome       string    `json:"outcome"`
}

type PaginatedAuditResponse struct {
	Content       []AuditEventResponse `json:"content"`
	Page          int                  `json:"page"`
	Size          int                  `json:"size"`
	TotalElements int64                `json:"totalElements"`
	TotalPages    int                  `json:"totalPages"`
}

type AuditDetailResponse struct {
	AuditEventResponse
	Journey []JourneyHopResponse `json:"journey,omitempty"`
}
