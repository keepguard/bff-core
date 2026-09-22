package http

import (
	"encoding/json"
	"net/http"
	"strings"

	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

type companyEntitlementSnapshot struct {
	Status        string `json:"status"`
	AllowsWrite   bool   `json:"allowsWrite"`
	AllowsIngest  bool   `json:"allowsIngest"`
	AllowsProduct bool   `json:"allowsProduct"`
}

// RequireBillingPlatformNotRestricted implementa o corte do Modo P:
// Se a organização estiver em estado restrito (inadimplência corporativa), bloqueia mutações (*:write)
// retornando HTTP 403 BILLING_RESTRICTED.
// Invariante da Spec: ADMIN da empresa NÃO bypassa; apenas ROLE_SYSTEM (ops global) bypassa.
func RequireBillingPlatformNotRestricted(billingClient domainclient.BillingClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if billingClient == nil {
				return next(c)
			}
			claims := GetClaimsFromContext(c)
			if claims == nil {
				return next(c)
			}
			// Bypass apenas para ROLE_SYSTEM (ops da KeepGuard)
			if pkg.HasAnyRole(claims.Roles, "SYSTEM") {
				return next(c)
			}

			companyID := domainclient.CompanyIDFromContext(c.Request().Context())
			if companyID == "" {
				companyID = strings.TrimSpace(claims.TenantId)
			}
			if companyID == "" {
				companyID = strings.TrimSpace(GetTenantId(c))
			}
			if companyID == "" {
				return next(c)
			}

			correlationID := GetCorrelationID(c)
			scope := domainclient.BillingScope{
				CompanyID:     companyID,
				UserID:        claims.UserID,
				CorrelationID: correlationID,
				Admin:         true,
			}

			raw, err := billingClient.GetCompanyEntitlement(c.Request().Context(), scope)
			if err != nil {
				// Falha transitória de comunicação com ms-billing não derruba a plataforma por fail-closed
				return next(c)
			}

			var snap companyEntitlementSnapshot
			if err := json.Unmarshal(raw, &snap); err == nil {
				status := strings.ToLower(strings.TrimSpace(snap.Status))
				if status == "restricted" || status == "canceled" || (!snap.AllowsWrite && status != "none" && status != "") {
					return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
						Error:         "BILLING_RESTRICTED",
						Message:       "Acesso de escrita suspenso por inadimplência da organização.",
						CorrelationID: correlationID,
					})
				}
			}

			return next(c)
		}
	}
}
