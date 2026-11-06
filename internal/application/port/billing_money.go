package port

import (
	"context"
	"encoding/json"
	"errors"
)

// BillingCall é uma chamada às rotas novas do ms-billing (Cobrança, fase C3). Sucesso é
// qualquer 2xx, com o status repassado.
type BillingCall struct {
	Operation string // rótulo de log e métrica
	Method    string
	Path      string // já com os parâmetros escapados
	Endpoint  string // caminho sem valores, para a métrica
	Query     map[string]string
	Body      any
}

// BillingResponse é a resposta 2xx de uma BillingCall.
type BillingResponse struct {
	Status int
	Body   json.RawMessage
}

// BillingForwarder repassa as rotas novas da Cobrança.
type BillingForwarder interface {
	Forward(ctx context.Context, scope BillingScope, call BillingCall) (BillingResponse, error)
}

// ActiveSession é uma entrada refreshtoken:<hash> do ms-auth.
type ActiveSession struct {
	SID       string
	CodeUser  string
	CompanyID string
	Revoked   bool
}

// ActiveSessionReader lê as chaves de sessão do ms-auth. Erro = indisponível, nunca
// "não existe": quem chama falha fechado.
type ActiveSessionReader interface {
	IsRevoked(ctx context.Context, sid string) (bool, error)
	Entries(ctx context.Context, sid string) ([]ActiveSession, error)
}

// UserAccess são os papéis da pessoa na company, lidos no banco do ms-auth.
type UserAccess struct {
	Active      bool
	Roles       []string
	Authorities []string
}

// ErrUserAccessRefused: o ms-auth não reconhece a pessoa na company, ou recusou o token.
var ErrUserAccessRefused = errors.New("acesso não reconhecido pelo ms-auth")

// UserAccessClient consulta os papéis na fonte. Devolve ErrUserAccessRefused quando o
// ms-auth responde que não; qualquer outro erro é indisponibilidade.
type UserAccessClient interface {
	GetUserAccess(ctx context.Context, codeUser, companyID, bearerToken, correlationID string) (UserAccess, error)
}
