package dto

type KnowledgeSourceHint struct {
	AgentID string `json:"agentId,omitempty"`
	Name    string `json:"name,omitempty"`
	Context string `json:"context,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
}

type KnowledgeAskRequest struct {
	Question    string                `json:"question"`
	Context     string                `json:"context,omitempty"`
	SourceHints []KnowledgeSourceHint `json:"sourceHints,omitempty"`
}

type KnowledgeAskSource struct {
	Kind          string `json:"kind"`
	Key           string `json:"key,omitempty"`
	DocumentID    string `json:"documentId,omitempty"`
	SourceAgentID string `json:"sourceAgentId,omitempty"`
	AgentName     string `json:"agentName,omitempty"`
	CollectedAt   string `json:"collectedAt,omitempty"`
	Excerpt       string `json:"excerpt,omitempty"`
}

type KnowledgeAskAudit struct {
	DocumentIDs []string `json:"documentIds"`
	Checks      []string `json:"checks"`
}

type KnowledgeFreshness struct {
	LastCollectionAt string `json:"lastCollectionAt,omitempty"`
	AgeMinutes       int    `json:"ageMinutes"`
	Status           string `json:"status"`
	Failed           bool   `json:"failed"`
	ErrorMessage     string `json:"errorMessage,omitempty"`
	AgentID          string `json:"agentId,omitempty"`
	AgentName        string `json:"agentName,omitempty"`
}

type KnowledgeAskResponse struct {
	Intent      string               `json:"intent"`
	Mode        string               `json:"mode"`
	Answer      string               `json:"answer"`
	ObservedAt  string               `json:"observedAt,omitempty"`
	Stale       bool                 `json:"stale"`
	Conflict    bool                 `json:"conflict"`
	Convergence bool                 `json:"convergence"`
	Unknown     bool                 `json:"unknown"`
	Sources     []KnowledgeAskSource `json:"sources"`
	AgeMinutes  *int64               `json:"ageMinutes,omitempty"`
	Audit       *KnowledgeAskAudit   `json:"audit,omitempty"`
	Disclaimer  string               `json:"disclaimer,omitempty"`
	Freshness   *KnowledgeFreshness  `json:"freshness,omitempty"`
}

type KnowledgeSnapshotDTO struct {
	ID            string         `json:"id"`
	CompanyID     string         `json:"companyId"`
	AgentID       string         `json:"agentId"`
	AgentCode     string         `json:"agentCode"`
	CollectorType string         `json:"collectorType"`
	Context       string         `json:"context"`
	EntityHint    string         `json:"entityHint"`
	CollectedAt   string         `json:"collectedAt"`
	IngestedAt    string         `json:"ingestedAt"`
	Schema        string         `json:"schema"`
	SourceURL     string         `json:"sourceUrl"`
	Payload       map[string]any `json:"payload"`
}

type KnowledgeDocumentPreviewDTO struct {
	ID               string `json:"id"`
	CompanyID        string `json:"companyId"`
	SourceAgentID    string `json:"sourceAgentId"`
	FileName         string `json:"fileName"`
	ContentType      string `json:"contentType"`
	CollectedAt      string `json:"collectedAt"`
	EntityHint       string `json:"entityHint"`
	Status           string `json:"status"`
	PreviewText      string `json:"previewText"`
	PreviewAvailable bool   `json:"previewAvailable"`
	Message          string `json:"message"`
}

type KnowledgeCollectionResultsDTO struct {
	Snapshots []KnowledgeSnapshotDTO        `json:"snapshots"`
	Documents []KnowledgeDocumentPreviewDTO `json:"documents"`
}

type ExecutionPayloadItemDTO struct {
	Kind        string         `json:"kind"`
	ID          string         `json:"id"`
	ContentType string         `json:"contentType,omitempty"`
	FileName    string         `json:"fileName,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	PreviewText string         `json:"previewText,omitempty"`
}
