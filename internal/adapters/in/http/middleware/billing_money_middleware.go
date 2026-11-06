package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	client "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	// AuthorityBillingAdmin é a authority de Operador da Cobrança (contrato C2 §3).
	AuthorityBillingAdmin = "billing:admin"

	coreWriteDisabledCode = "CORE_WRITE_DISABLED"
	sessionCheckTimeout   = 500 * time.Millisecond
)

// IsBillingOperator é a regra de Operador do contrato C2 §3: billing:admin, ROLE_ADMIN ou
// ROLE_SYSTEM. billing:write e ROLE_ORGANIZATION_ADMIN não bastam.
func IsBillingOperator(roles, authorities []string) bool {
	return pkg.HasAnyRole(roles, "ADMIN", "SYSTEM") || pkg.HasAuthority(authorities, AuthorityBillingAdmin)
}

// RequireBillingAdmin exige Operador, pelo que o token diz. Rotas de leitura de operador
// param aqui; escrita passa também pelo BillingMoneyGuard.
func RequireBillingAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaimsFromContext(c)
			if claims != nil && IsBillingOperator(claims.Roles, claims.Authorities) {
				return next(c)
			}
			return writeMoneyError(c, http.StatusForbidden, "FORBIDDEN", "Acesso restrito a quem opera a cobrança.")
		}
	}
}

// BillingMoneyGuard trava a escrita de Operador pelo bff-core (contrato C2 §3). A lista é
// explícita e, em produção, vazia: toda escrita recusa com 403 CORE_WRITE_DISABLED.
type BillingMoneyGuard struct {
	companies map[string]struct{}
	sessions  client.ActiveSessionReader
	access    client.UserAccessClient
	logger    *zap.Logger
}

// ParseCoreWriteCompanies lê a lista separada por vírgula. Só entra UUID: curinga e valor
// inválido saem em rejected. Texto vazio = lista vazia.
func ParseCoreWriteCompanies(raw string) (companies []string, rejected []string) {
	for _, item := range strings.Split(raw, ",") {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		if !isCanonicalUUID(item) {
			rejected = append(rejected, item)
			continue
		}
		companies = append(companies, item)
	}
	return companies, rejected
}

func isCanonicalUUID(value string) bool {
	if len(value) != 36 || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
			continue
		}
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// NewBillingMoneyGuard monta a trava. Sem a chave do billing configurada, a lista é
// ignorada: nada anexaria credencial, então tudo fica desligado.
func NewBillingMoneyGuard(companies []string, serviceKeyConfigured bool, sessions client.ActiveSessionReader, access client.UserAccessClient, logger *zap.Logger) *BillingMoneyGuard {
	if logger == nil {
		logger = zap.NewNop()
	}
	guard := &BillingMoneyGuard{companies: map[string]struct{}{}, sessions: sessions, access: access, logger: logger}
	if !serviceKeyConfigured || sessions == nil || access == nil {
		return guard
	}
	for _, company := range companies {
		guard.companies[strings.ToLower(strings.TrimSpace(company))] = struct{}{}
	}
	return guard
}

// Chain é a trava inteira, nesta ordem: Operador pelo token, lista de companies, sessão
// ativa e papéis na fonte. Nil vale como lista vazia.
func (g *BillingMoneyGuard) Chain() []echo.MiddlewareFunc {
	if g == nil {
		g = &BillingMoneyGuard{companies: map[string]struct{}{}, logger: zap.NewNop()}
	}
	return []echo.MiddlewareFunc{RequireBillingAdmin(), g.requireListedCompany(), g.requireActiveSession()}
}

func (g *BillingMoneyGuard) requireListedCompany() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			companyID := strings.ToLower(strings.TrimSpace(client.CompanyIDFromContext(c.Request().Context())))
			if _, listed := g.companies[companyID]; companyID != "" && listed {
				return next(c)
			}
			g.logger.Warn("Escrita de Operador recusada: company fora da lista",
				zap.String("correlationId", GetCorrelationID(c)),
				zap.String("companyId", companyID),
				zap.String("path", c.Path()),
			)
			return writeMoneyError(c, http.StatusForbidden, coreWriteDisabledCode, "Escrita desligada por configuração.")
		}
	}
}

// requireActiveSession é o RequireActiveSession do contrato C2 §3, sempre falhando fechado:
// Redis ou ms-auth fora do ar respondem 503, ao contrário da checagem do JWTMiddleware.
func (g *BillingMoneyGuard) requireActiveSession() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaimsFromContext(c)
			if claims == nil || strings.TrimSpace(claims.SID) == "" {
				return writeMoneyError(c, http.StatusUnauthorized, "SESSION_REQUIRED", "Sessão obrigatória nesta operação.")
			}
			sid := strings.TrimSpace(claims.SID)
			// Uma identidade só: quem opera (sub) tem de ser quem tem a sessão.
			codeUser := strings.TrimSpace(claims.Sub)
			if codeUser == "" || (strings.TrimSpace(claims.CodeUser) != "" && !strings.EqualFold(strings.TrimSpace(claims.CodeUser), codeUser)) {
				return writeMoneyError(c, http.StatusUnauthorized, "SESSION_INVALID", "Sessão não reconhecida.")
			}
			companyID := client.CompanyIDFromContext(c.Request().Context())

			ctx, cancel := context.WithTimeout(c.Request().Context(), sessionCheckTimeout)
			defer cancel()
			revoked, err := g.sessions.IsRevoked(ctx, sid)
			if err != nil {
				return g.unavailable(c, "redis", err)
			}
			if revoked {
				return writeMoneyError(c, http.StatusUnauthorized, "TOKEN_REVOKED", "Sessão revogada ou expirada.")
			}
			entries, err := g.sessions.Entries(ctx, sid)
			if err != nil {
				return g.unavailable(c, "redis", err)
			}
			if !hasLiveEntry(entries, sid, codeUser, companyID) {
				return writeMoneyError(c, http.StatusUnauthorized, "SESSION_INVALID", "Sessão não reconhecida.")
			}

			access, err := g.access.GetUserAccess(c.Request().Context(), codeUser, companyID, GetTokenFromContext(c), GetCorrelationID(c))
			if errors.Is(err, client.ErrUserAccessRefused) {
				return writeMoneyError(c, http.StatusForbidden, "FORBIDDEN", "Acesso restrito a quem opera a cobrança.")
			}
			if err != nil {
				return g.unavailable(c, "ms-auth", err)
			}
			if !access.Active || !IsBillingOperator(access.Roles, access.Authorities) {
				g.logger.Warn("Escrita de Operador recusada: papéis do token diferentes da fonte",
					zap.String("correlationId", GetCorrelationID(c)),
					zap.String("codeUser", codeUser),
					zap.String("path", c.Path()),
				)
				return writeMoneyError(c, http.StatusForbidden, "FORBIDDEN", "Acesso restrito a quem opera a cobrança.")
			}
			return next(c)
		}
	}
}

// hasLiveEntry é a prova positiva: uma entrada da sessão, sem revoked, da mesma pessoa
// e da company resolvida. Um sid inventado não tem entrada.
func hasLiveEntry(entries []client.ActiveSession, sid, codeUser, companyID string) bool {
	for _, entry := range entries {
		if entry.Revoked || entry.SID != sid {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(entry.CodeUser), codeUser) && strings.EqualFold(strings.TrimSpace(entry.CompanyID), companyID) {
			return true
		}
	}
	return false
}

func (g *BillingMoneyGuard) unavailable(c echo.Context, source string, err error) error {
	g.logger.Warn("Conferência de sessão indisponível: escrita recusada",
		zap.String("correlationId", GetCorrelationID(c)),
		zap.String("source", source),
		zap.String("path", c.Path()),
		zap.Error(err),
	)
	return writeMoneyError(c, http.StatusServiceUnavailable, "SESSION_CHECK_UNAVAILABLE", "Não foi possível conferir a sessão. Tente de novo em instantes.")
}

func writeMoneyError(c echo.Context, status int, code, message string) error {
	return c.JSON(status, pkg.ErrorResponse{Error: code, Message: message, CorrelationID: GetCorrelationID(c)})
}
