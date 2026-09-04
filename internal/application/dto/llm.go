package dto

type LlmUsageResponse struct {
	ID               string   `json:"id"`
	OccurredAt       string   `json:"occurredAt"`
	CompanyID        string   `json:"companyId,omitempty"`
	TenantID         string   `json:"tenantId,omitempty"`
	SourceService    string   `json:"sourceService,omitempty"`
	Feature          string   `json:"feature,omitempty"`
	ProviderID       string   `json:"providerId,omitempty"`
	ProviderType     string   `json:"providerType"`
	Model            string   `json:"model,omitempty"`
	PromptTokens     int      `json:"promptTokens"`
	CompletionTokens int      `json:"completionTokens"`
	TotalTokens      int      `json:"totalTokens"`
	EstimatedCostUSD *float64 `json:"estimatedCostUsd,omitempty"`
	Outcome          string   `json:"outcome"`
	LatencyMS        int      `json:"latencyMs"`
	CorrelationID    string   `json:"correlationId,omitempty"`
	RequestID        string   `json:"requestId,omitempty"`
	ErrorCode        string   `json:"errorCode,omitempty"`
}

type PaginatedLlmUsageResponse struct {
	Content       []LlmUsageResponse `json:"content"`
	Page          int                `json:"page"`
	Size          int                `json:"size"`
	TotalElements int64              `json:"totalElements"`
	TotalPages    int                `json:"totalPages"`
}
