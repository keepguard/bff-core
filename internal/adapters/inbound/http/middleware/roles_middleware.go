package http

import (
	"net/http"

	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

// RequireAnyRole exige JWT com ao menos uma das roles informadas.
func RequireAnyRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			claims := GetClaimsFromContext(c)
			if claims == nil || !pkg.HasAnyRole(claims.Roles, roles...) {
				return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
					Error:         "FORBIDDEN",
					Message:       "Acesso restrito a administradores",
					CorrelationID: correlationID,
				})
			}
			return next(c)
		}
	}
}

const (
	AuthorityAuditRead      = "audit:read"
	AuthorityKnowledgeRead  = "knowledge:read"
	AuthorityLlmRead        = "llm:read"
	AuthorityLlmWrite       = "llm:write"
	AuthorityCollectorRead  = "collector:read"
	AuthorityCollectorWrite = "collector:write"
	AuthorityGuardianRead   = "guardian:read"
	AuthorityGuardianWrite  = "guardian:write"
	AuthorityOAuthRead      = "oauth:read"
	AuthorityOAuthWrite     = "oauth:write"
	AuthorityOpsRead        = "ops:read"
	AuthorityBillingRead    = "billing:read"
	AuthorityBillingWrite   = "billing:write"
)

func RequireAuthority(authority string) echo.MiddlewareFunc {
	message := "Acesso restrito a administradores ou à permissão " + authority
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			claims := GetClaimsFromContext(c)
			if claims != nil && (pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") || pkg.HasAuthority(claims.Authorities, authority)) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
				Error:         "FORBIDDEN",
				Message:       message,
				CorrelationID: correlationID,
			})
		}
	}
}

func RequireKnowledgeRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityKnowledgeRead)
}

func RequireAuditRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityAuditRead)
}

func RequireLlmRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityLlmRead)
}

func RequireLlmWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityLlmWrite)
}

func RequireCollectorRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityCollectorRead)
}

func RequireCollectorWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityCollectorWrite)
}

func RequireGuardianRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityGuardianRead)
}

func RequireGuardianWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityGuardianWrite)
}

func RequireOAuthRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityOAuthRead)
}

func RequireOAuthWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityOAuthWrite)
}

func RequireOpsRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityOpsRead)
}

func RequireBillingRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityBillingRead)
}

func RequireBillingWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthorityBillingWrite)
}

func IsBillingOrgCaller(claims *pkg.JWTClaims) bool {
	if claims == nil {
		return false
	}
	if pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") {
		return true
	}
	if !pkg.HasAnyRole(claims.Roles, "MANAGER") {
		return false
	}
	return pkg.HasAuthority(claims.Authorities, AuthorityBillingRead) || pkg.HasAuthority(claims.Authorities, AuthorityBillingWrite)
}

func RequireBillingOrgRead() echo.MiddlewareFunc {
	message := "Acesso restrito a administradores ou gestores com billing:read"
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			if IsBillingOrgCaller(GetClaimsFromContext(c)) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
				Error:         "FORBIDDEN",
				Message:       message,
				CorrelationID: correlationID,
			})
		}
	}
}
