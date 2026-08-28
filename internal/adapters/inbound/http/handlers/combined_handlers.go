package handlers

// CombinedHandlers une handlers públicos de registro e o perfil autenticado.
type CombinedHandlers struct {
	*RegisterHandlers
	*UserHandlers
}

// NewCombinedHandlers cria CombinedHandlers.
func NewCombinedHandlers(registerHandlers *RegisterHandlers, userHandlers *UserHandlers) *CombinedHandlers {
	return &CombinedHandlers{
		RegisterHandlers: registerHandlers,
		UserHandlers:     userHandlers,
	}
}
