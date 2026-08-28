package http

import (
	"net/http"
	"strings"

	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

// CompanyResolveMiddleware resolve tenant (JWT ou X-Tenant-Id) → company e anexa companyId ao contexto.
func CompanyResolveMiddleware(companyClient client.CompanyClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger") {
				return next(c)
			}

			tenantId := tenantIDFromRequest(c)
			if tenantId == "" {
				return next(c)
			}

			correlationID := GetCorrelationID(c)
			company, err := companyClient.GetByTenantId(c.Request().Context(), tenantId, correlationID)
			if err != nil {
				return err
			}
			if company.ID == "" {
				return echo.NewHTTPError(http.StatusNotFound, "Empresa não encontrada para o tenant informado")
			}

			req := c.Request().WithContext(client.WithCompanyID(c.Request().Context(), company.ID))
			c.SetRequest(req)
			return next(c)
		}
	}
}

func tenantIDFromRequest(c echo.Context) string {
	headerTenant := strings.TrimSpace(c.Request().Header.Get("X-Tenant-Id"))
	if claims := GetClaimsFromContext(c); claims != nil && claims.TenantId != "" {
		return claims.TenantId
	}
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		if claims, err := pkg.ExtractAllClaims(authHeader); err == nil && claims != nil && strings.TrimSpace(claims.TenantId) != "" {
			return strings.TrimSpace(claims.TenantId)
		}
	}
	return headerTenant
}
