package handlers

// CombinedHandlers une handlers públicos de registro, perfil autenticado e consentimentos.
type CombinedHandlers struct {
	*RegisterHandlers
	*UserHandlers
	*ConsentHandlers
	*ConnectionsHandlers
	*AuditHandlers
	*GuardianHandlers
	*OAuthClientHandlers
	*CollectorAgentHandlers
	*KnowledgeHandlers
	*LlmHandlers
	*BillingHandlers
}

// NewCombinedHandlers cria CombinedHandlers.
func NewCombinedHandlers(
	registerHandlers *RegisterHandlers,
	userHandlers *UserHandlers,
	consentHandlers *ConsentHandlers,
	connectionsHandlers *ConnectionsHandlers,
	auditHandlers *AuditHandlers,
	guardianHandlers *GuardianHandlers,
	oauthClientHandlers *OAuthClientHandlers,
	collectorAgentHandlers *CollectorAgentHandlers,
	knowledgeHandlers *KnowledgeHandlers,
	llmHandlers *LlmHandlers,
	billingHandlers *BillingHandlers,
) *CombinedHandlers {
	return &CombinedHandlers{
		RegisterHandlers:       registerHandlers,
		UserHandlers:           userHandlers,
		ConsentHandlers:        consentHandlers,
		ConnectionsHandlers:    connectionsHandlers,
		AuditHandlers:          auditHandlers,
		GuardianHandlers:       guardianHandlers,
		OAuthClientHandlers:    oauthClientHandlers,
		CollectorAgentHandlers: collectorAgentHandlers,
		KnowledgeHandlers:      knowledgeHandlers,
		LlmHandlers:            llmHandlers,
		BillingHandlers:        billingHandlers,
	}
}
