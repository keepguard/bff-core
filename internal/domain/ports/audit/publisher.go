package audit

import "context"

type Actor struct {
	Type     string   `json:"type"`
	CodeUser string   `json:"codeUser,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	ClientIP string   `json:"clientIp,omitempty"`
	DeviceID string   `json:"deviceId,omitempty"`
}

type Resource struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type Event struct {
	EventID       string         `json:"eventId"`
	OccurredAt    string         `json:"occurredAt"`
	SchemaVersion int            `json:"schemaVersion"`
	SourceService string         `json:"sourceService"`
	CorrelationID string         `json:"correlationId"`
	RequestID     string         `json:"requestId,omitempty"`
	TenantID      string         `json:"tenantId,omitempty"`
	CompanyID     string         `json:"companyId,omitempty"`
	Actor         Actor          `json:"actor"`
	Action        string         `json:"action"`
	Resource      Resource       `json:"resource"`
	Outcome       string         `json:"outcome"`
	Reason        string         `json:"reason,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event)
	Close() error
}
