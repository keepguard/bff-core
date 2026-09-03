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

const AuthorityAuditRead = "audit:read"
const AuthorityKnowledgeRead = "knowledge:read"

// RequireKnowledgeRead permite ADMIN, SYSTEM ou a authority knowledge:read.
func RequireKnowledgeRead() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			claims := GetClaimsFromContext(c)
			if claims != nil && (pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") || pkg.HasAuthority(claims.Authorities, AuthorityKnowledgeRead)) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
				Error:         "FORBIDDEN",
				Message:       "Acesso restrito a administradores ou à permissão knowledge:read",
				CorrelationID: correlationID,
			})
		}
	}
}

// RequireAuditRead permite ADMIN, SYSTEM ou a authority audit:read.
func RequireAuditRead() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := GetCorrelationID(c)
			claims := GetClaimsFromContext(c)
			if claims != nil && (pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") || pkg.HasAuthority(claims.Authorities, AuthorityAuditRead)) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
				Error:         "FORBIDDEN",
				Message:       "Acesso restrito a administradores ou à permissão audit:read",
				CorrelationID: correlationID,
			})
		}
	}
}
