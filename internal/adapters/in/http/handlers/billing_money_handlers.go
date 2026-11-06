package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/in/http/middleware"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/clientip"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
)

const (
	billingAPIBase      = "/api/v1/billing"
	billingMoneyMaxBody = 64 * 1024
)

// GetBillingCreditsHandler é GET /credits. userId só vai adiante para Operador: quem não
// opera lê o próprio extrato (contrato C2 §3 e §6.12).
func (h *BillingHandlers) GetBillingCreditsHandler(c echo.Context) error {
	query := billingListQuery(c, "page", "size", "userId")
	claims := middlewarePkg.GetClaimsFromContext(c)
	if claims == nil || !middlewarePkg.IsBillingOperator(claims.Roles, claims.Authorities) {
		delete(query, "userId")
	}
	return h.forward(c, port.BillingCall{Operation: "GetCredits", Method: http.MethodGet, Path: billingAPIBase + "/credits", Endpoint: "/billing/credits", Query: query})
}

// RemoveBillingCardHandler é DELETE /subscriptions/:id/card. Só o pagador: a regra é do billing.
func (h *BillingHandlers) RemoveBillingCardHandler(c echo.Context) error {
	return h.forward(c, port.BillingCall{
		Operation: "RemoveCard", Method: http.MethodDelete,
		Path: billingAPIBase + "/subscriptions/" + url.PathEscape(c.Param("id")) + "/card", Endpoint: "/billing/subscriptions/{id}/card",
	})
}

// GetBillingSettingsHandler é GET /billing-settings, de Operador.
func (h *BillingHandlers) GetBillingSettingsHandler(c echo.Context) error {
	return h.forward(c, port.BillingCall{Operation: "GetBillingSettings", Method: http.MethodGet, Path: billingAPIBase + "/billing-settings", Endpoint: "/billing/billing-settings"})
}

// ListBillingReviewsHandler é GET /billing-reviews, de Operador.
func (h *BillingHandlers) ListBillingReviewsHandler(c echo.Context) error {
	return h.forward(c, port.BillingCall{
		Operation: "ListBillingReviews", Method: http.MethodGet, Path: billingAPIBase + "/billing-reviews", Endpoint: "/billing/billing-reviews",
		Query: billingListQuery(c, "page", "size", "status"),
	})
}

// GetBillingRefundPreviewHandler é GET /invoices/:id/refund-preview, de Operador.
func (h *BillingHandlers) GetBillingRefundPreviewHandler(c echo.Context) error {
	return h.forward(c, port.BillingCall{
		Operation: "GetRefundPreview", Method: http.MethodGet,
		Path: billingAPIBase + "/invoices/" + url.PathEscape(c.Param("id")) + "/refund-preview", Endpoint: "/billing/invoices/{id}/refund-preview",
		Query: billingListQuery(c, "amountCents", "method"),
	})
}

// As rotas abaixo ficam atrás do BillingMoneyGuard (contrato C2 §3).

// GrantBillingCreditHandler é POST /credits.
func (h *BillingHandlers) GrantBillingCreditHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{Operation: "GrantCredit", Method: http.MethodPost, Path: billingAPIBase + "/credits", Endpoint: "/billing/credits"})
}

// ReverseBillingCreditHandler é POST /credits/:id/reverse.
func (h *BillingHandlers) ReverseBillingCreditHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{
		Operation: "ReverseCredit", Method: http.MethodPost,
		Path: billingAPIBase + "/credits/" + url.PathEscape(c.Param("id")) + "/reverse", Endpoint: "/billing/credits/{id}/reverse",
	})
}

// RefundBillingInvoiceHandler é POST /invoices/:id/refunds.
func (h *BillingHandlers) RefundBillingInvoiceHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{
		Operation: "RefundInvoice", Method: http.MethodPost,
		Path: billingAPIBase + "/invoices/" + url.PathEscape(c.Param("id")) + "/refunds", Endpoint: "/billing/invoices/{id}/refunds",
	})
}

// ReleaseBillingRestrictionHandler é POST /entitlements/:userId/release-restriction.
func (h *BillingHandlers) ReleaseBillingRestrictionHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{
		Operation: "ReleaseRestriction", Method: http.MethodPost,
		Path: billingAPIBase + "/entitlements/" + url.PathEscape(c.Param("userId")) + "/release-restriction", Endpoint: "/billing/entitlements/{userId}/release-restriction",
	})
}

// PutBillingSettingsHandler é PUT /billing-settings. Esta chave nunca liga company nem muda
// piso e tetos: recusa aqui, sem depender só do billing (que também recusa).
func (h *BillingHandlers) PutBillingSettingsHandler(c echo.Context) error {
	body, err := bindBillingMoneyJSON(c)
	if err != nil || c.Response().Committed {
		return err
	}
	if touchesCoreForbiddenSettings(body) {
		return handleError(c, pkg.NewAppError("CORE_KEY_NOT_ALLOWED", "Esta credencial não liga a company nem muda piso e tetos.", http.StatusForbidden), middlewarePkg.GetCorrelationID(c))
	}
	return h.forwardOperator(c, port.BillingCall{Operation: "PutBillingSettings", Method: http.MethodPut, Path: billingAPIBase + "/billing-settings", Endpoint: "/billing/billing-settings", Body: body})
}

// AddBillingTestAccountHandler é POST /test-accounts.
func (h *BillingHandlers) AddBillingTestAccountHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{Operation: "AddTestAccount", Method: http.MethodPost, Path: billingAPIBase + "/test-accounts", Endpoint: "/billing/test-accounts"})
}

// RemoveBillingTestAccountHandler é DELETE /test-accounts?userId=.
func (h *BillingHandlers) RemoveBillingTestAccountHandler(c echo.Context) error {
	return h.forwardOperator(c, port.BillingCall{
		Operation: "RemoveTestAccount", Method: http.MethodDelete, Path: billingAPIBase + "/test-accounts", Endpoint: "/billing/test-accounts",
		Query: billingListQuery(c, "userId"),
	})
}

// ResolveBillingReviewHandler é POST /billing-reviews/:id/resolve.
func (h *BillingHandlers) ResolveBillingReviewHandler(c echo.Context) error {
	return h.forwardMoney(c, port.BillingCall{
		Operation: "ResolveBillingReview", Method: http.MethodPost,
		Path: billingAPIBase + "/billing-reviews/" + url.PathEscape(c.Param("id")) + "/resolve", Endpoint: "/billing/billing-reviews/{id}/resolve",
	})
}

// forward repassa uma rota nova com a identidade de quem chama.
func (h *BillingHandlers) forward(c echo.Context, call port.BillingCall) error {
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	return h.send(c, scope, call)
}

// forwardMoney lê o corpo, recusa dado de cartão e repassa como Operador.
func (h *BillingHandlers) forwardMoney(c echo.Context, call port.BillingCall) error {
	body, err := bindBillingMoneyJSON(c)
	if err != nil || c.Response().Committed {
		return err
	}
	call.Body = body
	return h.forwardOperator(c, call)
}

// forwardOperator anexa a credencial do bff-core, quem opera (sub do token) e o IP real.
func (h *BillingHandlers) forwardOperator(c echo.Context, call port.BillingCall) error {
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	claims := middlewarePkg.GetClaimsFromContext(c)
	if claims != nil {
		scope.OperatorID = strings.TrimSpace(claims.Sub)
		if scope.OperatorID == "" {
			scope.OperatorID = strings.TrimSpace(claims.CodeUser)
		}
	}
	scope.ClientIP = clientip.FromRequest(c.Request())
	return h.send(c, scope, call)
}

func (h *BillingHandlers) send(c echo.Context, scope port.BillingScope, call port.BillingCall) error {
	resp, err := h.billing.Forward(c.Request().Context(), scope, call)
	if err != nil {
		return handleError(c, err, scope.CorrelationID)
	}
	if resp.Status == http.StatusNoContent {
		return c.NoContent(http.StatusNoContent)
	}
	return writeRaw(c, resp.Status, resp.Body)
}

// bindBillingMoneyJSON lê o corpo de escrita de Operador: número intacto, no tamanho
// máximo e sem dado de cartão. Quando recusa, já escreveu a resposta.
func bindBillingMoneyJSON(c echo.Context) (map[string]any, error) {
	correlationID := middlewarePkg.GetCorrelationID(c)
	raw, err := io.ReadAll(io.LimitReader(c.Request().Body, billingMoneyMaxBody+1))
	if err != nil || len(raw) > billingMoneyMaxBody {
		return nil, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID})
	}
	body := map[string]any{}
	if len(bytes.TrimSpace(raw)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil {
			return nil, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{Error: "BAD_REQUEST", Message: "Payload inválido", CorrelationID: correlationID})
		}
	}
	if containsCardData(body) {
		return nil, c.JSON(http.StatusUnprocessableEntity, pkg.ErrorResponse{Error: "PAYMENT_METHOD_UNSUPPORTED", Message: "PAN/CVV não são aceitos", CorrelationID: correlationID})
	}
	return body, nil
}

// touchesCoreForbiddenSettings: ligar a company ou mexer em piso e tetos. A chave é
// normalizada porque o decoder do billing não diferencia maiúsculas.
func touchesCoreForbiddenSettings(body map[string]any) bool {
	for key, value := range body {
		switch normalizeBillingKey(key) {
		case "minchargecents", "creditcapbp", "dailycreditcapcents", "dailyrefundcapcents":
			return true
		case "financialmode":
			if strings.EqualFold(strings.TrimSpace(stringifyJSON(value)), "ON") {
				return true
			}
		}
	}
	return false
}
