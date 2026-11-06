package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	client "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

const (
	moneyCompany = "4e65bc72-c60b-4382-977d-c15314d117a8"
	moneyUser    = "1a2b3c4d-0000-4000-8000-000000000001"
	moneySID     = "sid-real"
)

type fakeSessions struct {
	revoked    bool
	entries    []client.ActiveSession
	revokedErr error
	entriesErr error
}

func (f *fakeSessions) IsRevoked(context.Context, string) (bool, error) {
	return f.revoked, f.revokedErr
}
func (f *fakeSessions) Entries(context.Context, string) ([]client.ActiveSession, error) {
	return f.entries, f.entriesErr
}

type fakeUserAccess struct {
	access client.UserAccess
	err    error
	calls  int
}

func (f *fakeUserAccess) GetUserAccess(context.Context, string, string, string, string) (client.UserAccess, error) {
	f.calls++
	return f.access, f.err
}

func liveEntry() client.ActiveSession {
	return client.ActiveSession{SID: moneySID, CodeUser: moneyUser, CompanyID: moneyCompany}
}

func adminAccess() client.UserAccess {
	return client.UserAccess{Active: true, Roles: []string{"ROLE_ADMIN"}}
}

// runMoneyChain roda a corrente inteira e diz o status, o código e se o handler foi chamado.
func runMoneyChain(guard *BillingMoneyGuard, claims *pkg.JWTClaims, company string) (int, string, bool) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/credits", nil)
	if company != "" {
		req = req.WithContext(client.WithCompanyID(req.Context(), company))
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if claims != nil {
		c.Set("claims", claims)
		c.Set("token", "jwt-de-teste")
	}
	called := false
	handler := echo.HandlerFunc(func(echo.Context) error {
		called = true
		return nil
	})
	chain := guard.Chain()
	for i := len(chain) - 1; i >= 0; i-- {
		handler = chain[i](handler)
	}
	_ = handler(c)
	return rec.Code, rec.Body.String(), called
}

func adminClaims() *pkg.JWTClaims {
	return &pkg.JWTClaims{Sub: moneyUser, SID: moneySID, Roles: []string{"ROLE_ADMIN"}}
}

func TestParseCoreWriteCompanies(t *testing.T) {
	companies, rejected := ParseCoreWriteCompanies("")
	if len(companies) != 0 || len(rejected) != 0 {
		t.Fatalf("vazio: %v %v", companies, rejected)
	}
	companies, rejected = ParseCoreWriteCompanies(" " + strings.ToUpper(moneyCompany) + " , *, nao-uuid,00000000-0000-0000-0000-000000000000")
	if len(companies) != 1 || companies[0] != moneyCompany {
		t.Fatalf("companies %v", companies)
	}
	if len(rejected) != 3 {
		t.Fatalf("curinga, texto e zero ficam fora: %v", rejected)
	}
}

// K2: billing:write e ROLE_ORGANIZATION_ADMIN não bastam em dinheiro.
func TestIsBillingOperator(t *testing.T) {
	cases := []struct {
		name        string
		roles       []string
		authorities []string
		want        bool
	}{
		{"billing:admin", []string{"ROLE_USER"}, []string{"billing:admin"}, true},
		{"ROLE_ADMIN", []string{"ROLE_ADMIN"}, nil, true},
		{"ROLE_SYSTEM", []string{"ROLE_SYSTEM"}, nil, true},
		{"billing:write", []string{"ROLE_USER"}, []string{"billing:write"}, false},
		{"ROLE_ORGANIZATION_ADMIN", []string{"ROLE_ORGANIZATION_ADMIN"}, nil, false},
		{"ROLE_MANAGER com billing:read", []string{"ROLE_MANAGER"}, []string{"billing:read"}, false},
		{"billing:self-write", []string{"ROLE_USER"}, []string{"billing:self-write"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsBillingOperator(tc.roles, tc.authorities); got != tc.want {
				t.Fatalf("got %v", got)
			}
		})
	}
}

// K7: sem lista, ou sem chave, ou sem as portas, tudo recusa. Nil vale como lista vazia.
func TestMoneyGuardEmptyListRefuses(t *testing.T) {
	sessions := &fakeSessions{entries: []client.ActiveSession{liveEntry()}}
	access := &fakeUserAccess{access: adminAccess()}
	guards := map[string]*BillingMoneyGuard{
		"guard nil":     nil,
		"lista vazia":   NewBillingMoneyGuard(nil, true, sessions, access, nil),
		"sem chave":     NewBillingMoneyGuard([]string{moneyCompany}, false, sessions, access, nil),
		"sem sessoes":   NewBillingMoneyGuard([]string{moneyCompany}, true, nil, access, nil),
		"sem ms-auth":   NewBillingMoneyGuard([]string{moneyCompany}, true, sessions, nil, nil),
		"outra company": NewBillingMoneyGuard([]string{"aaaaaaaa-0000-4000-8000-000000000001"}, true, sessions, access, nil),
	}
	for name, guard := range guards {
		t.Run(name, func(t *testing.T) {
			status, body, called := runMoneyChain(guard, adminClaims(), moneyCompany)
			if status != http.StatusForbidden || !strings.Contains(body, "CORE_WRITE_DISABLED") || called {
				t.Fatalf("status %d body %s handler %v", status, body, called)
			}
		})
	}
	if access.calls != 0 {
		t.Fatalf("lista vazia não consulta o ms-auth: %d", access.calls)
	}
}

// K7.1: com a lista preenchida, sessão real e papéis na fonte; fora do ar, 503.
func TestMoneyGuardActiveSession(t *testing.T) {
	otherUser := liveEntry()
	otherUser.CodeUser = "1a2b3c4d-0000-4000-8000-000000000002"
	otherCompany := liveEntry()
	otherCompany.CompanyID = "aaaaaaaa-0000-4000-8000-000000000001"
	revokedEntry := liveEntry()
	revokedEntry.Revoked = true
	otherSID := liveEntry()
	otherSID.SID = "outra"
	boom := errors.New("fora do ar")

	cases := []struct {
		name     string
		claims   *pkg.JWTClaims
		sessions *fakeSessions
		access   *fakeUserAccess
		status   int
		code     string
	}{
		{name: "sessao real de quem tem a role na fonte", claims: adminClaims(), status: http.StatusOK},
		{name: "billing:admin na fonte e no token", claims: &pkg.JWTClaims{Sub: moneyUser, SID: moneySID, Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:admin"}},
			access: &fakeUserAccess{access: client.UserAccess{Active: true, Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:admin"}}}, status: http.StatusOK},
		{name: "sem sid", claims: &pkg.JWTClaims{Sub: moneyUser, Roles: []string{"ROLE_ADMIN"}}, status: http.StatusUnauthorized, code: "SESSION_REQUIRED"},
		{name: "token de servico nao tem sid", claims: &pkg.JWTClaims{Sub: moneyUser, Roles: []string{"ROLE_SYSTEM"}}, status: http.StatusUnauthorized, code: "SESSION_REQUIRED"},
		{name: "sid inventado", claims: adminClaims(), sessions: &fakeSessions{}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "sessao de outro usuario", claims: adminClaims(), sessions: &fakeSessions{entries: []client.ActiveSession{otherUser}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "sessao de outra company", claims: adminClaims(), sessions: &fakeSessions{entries: []client.ActiveSession{otherCompany}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "entrada revogada", claims: adminClaims(), sessions: &fakeSessions{entries: []client.ActiveSession{revokedEntry}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "entrada de outro sid", claims: adminClaims(), sessions: &fakeSessions{entries: []client.ActiveSession{otherSID}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "codeUser de outra pessoa no token", claims: &pkg.JWTClaims{Sub: "1a2b3c4d-0000-4000-8000-000000000002", CodeUser: moneyUser, SID: moneySID, Roles: []string{"ROLE_ADMIN"}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "sem sub", claims: &pkg.JWTClaims{CodeUser: moneyUser, SID: moneySID, Roles: []string{"ROLE_ADMIN"}}, status: http.StatusUnauthorized, code: "SESSION_INVALID"},
		{name: "session revoked", claims: adminClaims(), sessions: &fakeSessions{revoked: true, entries: []client.ActiveSession{liveEntry()}}, status: http.StatusUnauthorized, code: "TOKEN_REVOKED"},
		{name: "ROLE_ADMIN forjado sobre sessao real de ROLE_USER", claims: adminClaims(),
			access: &fakeUserAccess{access: client.UserAccess{Active: true, Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:read"}}}, status: http.StatusForbidden, code: "FORBIDDEN"},
		{name: "conta inativa na fonte", claims: adminClaims(), access: &fakeUserAccess{access: client.UserAccess{Active: false, Roles: []string{"ROLE_ADMIN"}}}, status: http.StatusForbidden, code: "FORBIDDEN"},
		{name: "fora da company no ms-auth", claims: adminClaims(), access: &fakeUserAccess{err: client.ErrUserAccessRefused}, status: http.StatusForbidden, code: "FORBIDDEN"},
		{name: "Redis fora na revogacao", claims: adminClaims(), sessions: &fakeSessions{revokedErr: boom}, status: http.StatusServiceUnavailable, code: "SESSION_CHECK_UNAVAILABLE"},
		{name: "Redis fora nas entradas", claims: adminClaims(), sessions: &fakeSessions{entriesErr: boom}, status: http.StatusServiceUnavailable, code: "SESSION_CHECK_UNAVAILABLE"},
		{name: "ms-auth fora", claims: adminClaims(), access: &fakeUserAccess{err: boom}, status: http.StatusServiceUnavailable, code: "SESSION_CHECK_UNAVAILABLE"},
		{name: "billing:write nao passa", claims: &pkg.JWTClaims{Sub: moneyUser, SID: moneySID, Roles: []string{"ROLE_USER"}, Authorities: []string{"billing:write"}}, status: http.StatusForbidden, code: "FORBIDDEN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessions := tc.sessions
			if sessions == nil {
				sessions = &fakeSessions{entries: []client.ActiveSession{liveEntry()}}
			}
			access := tc.access
			if access == nil {
				access = &fakeUserAccess{access: adminAccess()}
			}
			guard := NewBillingMoneyGuard([]string{moneyCompany}, true, sessions, access, nil)
			status, body, called := runMoneyChain(guard, tc.claims, moneyCompany)
			if status != tc.status {
				t.Fatalf("status %d body %s", status, body)
			}
			if tc.code != "" && !strings.Contains(body, tc.code) {
				t.Fatalf("esperado %s, veio %s", tc.code, body)
			}
			if called != (tc.status == http.StatusOK) {
				t.Fatalf("handler chamado = %v, status %d", called, status)
			}
		})
	}
}
