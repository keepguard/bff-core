package handlers

import (
	"net/http"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/labstack/echo/v4"
)

// GetBillingPaymentOptionsHandler é GET /core/billing/payment-options (stripe-contratos S0a §6): moeda,
// região, gateway e formas de pagar do país da conta. O país vem do ms-user, como em assinar e trocar
// (X-Payer-Country do navegador é descartado); sem conseguir ler, 503 PAYER_COUNTRY_UNAVAILABLE e o
// ms-billing não é chamado. Leitura: o decorator repete em falha de rede.
func (h *BillingHandlers) GetBillingPaymentOptionsHandler(c echo.Context) error {
	scope, ok, respErr := h.guard(c)
	if !ok {
		return respErr
	}
	if ok, respErr := h.attachPayerCountry(c, &scope); !ok {
		return respErr
	}
	return h.send(c, scope, port.BillingCall{
		Operation: "GetPaymentOptions", Method: http.MethodGet, Path: billingAPIBase + "/payment-options", Endpoint: "/billing/payment-options",
	})
}

// GetBillingLockStateHandler é GET /core/billing/lock-state (S0a §6): assinatura viva e saldo por
// moeda da própria pessoa (X-User-Id = sub do JWT).
func (h *BillingHandlers) GetBillingLockStateHandler(c echo.Context) error {
	return h.forward(c, port.BillingCall{
		Operation: "GetLockState", Method: http.MethodGet, Path: billingAPIBase + "/lock-state", Endpoint: "/billing/lock-state",
	})
}
