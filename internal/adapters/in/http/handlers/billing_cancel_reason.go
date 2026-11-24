package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/in/http/middleware"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
)

// cancelBodyMaxBytes é o maior corpo aceito no cancelar: o motivo (código + até 300 caracteres) cabe com folga.
const cancelBodyMaxBytes = 4 * 1024

// cancellationReasonFromBody lê o corpo do cancelar e devolve só o objeto cancellationReason, cru (o ms-billing
// valida código e texto). Corpo maior que o teto: 413, já respondido (ok=false). Sem corpo, corpo que não é
// um objeto JSON, ou sem o campo (ou null): nil, e o cancelar segue como antes. O corpo nunca vai a log.
func cancellationReasonFromBody(c echo.Context) (json.RawMessage, bool, error) {
	body := c.Request().Body
	if body == nil {
		return nil, true, nil
	}
	raw, err := io.ReadAll(io.LimitReader(body, cancelBodyMaxBytes+1))
	if err != nil {
		return nil, true, nil
	}
	if len(raw) > cancelBodyMaxBytes {
		return nil, false, c.JSON(http.StatusRequestEntityTooLarge, pkg.ErrorResponse{
			Error:         "PAYLOAD_TOO_LARGE",
			Message:       "Corpo da requisição grande demais",
			CorrelationID: middlewarePkg.GetCorrelationID(c),
		})
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, true, nil
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, true, nil
	}
	reason, ok := envelope["cancellationReason"]
	if !ok || bytes.Equal(bytes.TrimSpace(reason), []byte("null")) {
		return nil, true, nil
	}
	return reason, true, nil
}

// cancelWithReason é o cancelar com o corpo {"cancellationReason": ...}: mesmos cabeçalhos, Idempotency-Key,
// ator e regra do immediate do caminho de sempre.
func (h *BillingHandlers) cancelWithReason(c echo.Context, id string, immediate bool, reason json.RawMessage) error {
	scope, ok, respErr := h.guard(c)
	if !ok {
		return respErr
	}
	call := port.BillingCall{
		Operation: "CancelSubscription",
		Method:    http.MethodPost,
		Path:      billingAPIBase + "/subscriptions/" + url.PathEscape(id) + "/cancel",
		Endpoint:  "/billing/subscriptions/{id}/cancel",
		Body:      map[string]json.RawMessage{"cancellationReason": reason},
	}
	if immediate && scope.Admin {
		call.Query = map[string]string{"immediate": "true"}
	}
	return h.send(c, scope, call)
}
