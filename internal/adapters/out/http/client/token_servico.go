package client

import "github.com/keepguard/lib-go-common/servicetoken"

// tokenServico é o token de serviço do bff-core (OAuth client_credentials no ms-auth), da lib-go-common/servicetoken.
type tokenServico interface {
	Token() (string, error)
}

// tokenServicoDoAmbiente lê KEEPGUARD_OAUTH_*. Sem client configurado devolve nil de verdade (um *Source nil dentro
// da interface não seria nil e o "oauth != nil" chamaria Token num ponteiro nulo).
func tokenServicoDoAmbiente() tokenServico {
	if s := servicetoken.FromEnv(); s != nil {
		return s
	}
	return nil
}
