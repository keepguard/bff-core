package handlers

// CombinedHandlers une handlers públicos de registro, perfil autenticado e consentimentos.
type CombinedHandlers struct {
	*RegisterHandlers
	*UserHandlers
	*ConsentHandlers
	*ConnectionsHandlers
}

// NewCombinedHandlers cria CombinedHandlers.
func NewCombinedHandlers(registerHandlers *RegisterHandlers, userHandlers *UserHandlers, consentHandlers *ConsentHandlers, connectionsHandlers *ConnectionsHandlers) *CombinedHandlers {
	return &CombinedHandlers{
		RegisterHandlers:    registerHandlers,
		UserHandlers:        userHandlers,
		ConsentHandlers:     consentHandlers,
		ConnectionsHandlers: connectionsHandlers,
	}
}
