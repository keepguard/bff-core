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
