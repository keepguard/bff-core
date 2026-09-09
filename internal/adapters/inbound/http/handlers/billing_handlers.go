package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const asaasWebhookMaxBytes = 512 * 1024

type BillingHandlers struct {
	billing billing.BillingPort
	users   port.UserClient
	logger  *zap.Logger
}

func NewBillingHandlers(billingPort billing.BillingPort, logger *zap.Logger) *BillingHandlers {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BillingHandlers{billing: billingPort, logger: logger}
}

func (h *BillingHandlers) WithUsers(users port.UserClient) *BillingHandlers {
	h.users = users
	return h
}

func (h *BillingHandlers) GetBillingEntitlementHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.GetEntitlement(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) ListBillingPlansHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.ListPlans(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) CreateBillingPlanHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	return h.proxy(c, http.StatusCreated, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.SavePlan(ctx.Request().Context(), scope, body)
	})
}

func (h *BillingHandlers) PatchBillingPlanHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	code := c.Param("code")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.PatchPlan(ctx.Request().Context(), scope, code, body)
	})
}

func (h *BillingHandlers) GetBillingGatewayAccountHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.GetGatewayAccount(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) PutBillingGatewayAccountHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.PutGatewayAccount(ctx.Request().Context(), scope, body)
	})
}

func (h *BillingHandlers) GetBillingSubscriptionHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.GetSubscription(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) CreateBillingSubscriptionHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	if scope.Admin {
		return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
			Error:         "BILLING_PAYER_OPS",
			Message:       "Quem opera a organização não assina",
			CorrelationID: scope.CorrelationID,
		})
	}
	if attachErr := h.attachPayerProfile(c, body, scope); attachErr != nil {
		return attachErr
	}
	result, status, callErr := h.billing.CreateSubscription(c.Request().Context(), scope, body)
	if callErr != nil {
		return handleError(c, callErr, scope.CorrelationID)
	}
	if status == 0 {
		status = http.StatusCreated
	}
	return writeRaw(c, status, result)
}

func (h *BillingHandlers) CancelBillingSubscriptionHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.CancelSubscription(ctx.Request().Context(), scope, id)
	})
}

func (h *BillingHandlers) ListBillingInvoicesHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.ListInvoices(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) GetBillingInvoiceHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.GetInvoice(ctx.Request().Context(), scope, id)
	})
}

func (h *BillingHandlers) AsaasWebhookHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if c.QueryString() != "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "query string recusada no webhook",
			CorrelationID: correlationID,
		})
	}
	token := strings.TrimSpace(c.Request().Header.Get("asaas-access-token"))
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "webhook token inválido",
			CorrelationID: correlationID,
		})
	}
	if c.Request().ContentLength > asaasWebhookMaxBytes {
		return c.JSON(http.StatusRequestEntityTooLarge, pkg.ErrorResponse{
			Error:         "PAYLOAD_TOO_LARGE",
			Message:       "body excede o teto",
			CorrelationID: correlationID,
		})
	}
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, asaasWebhookMaxBytes+1))
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "body inválido",
			CorrelationID: correlationID,
		})
	}
	if len(body) > asaasWebhookMaxBytes {
		return c.JSON(http.StatusRequestEntityTooLarge, pkg.ErrorResponse{
			Error:         "PAYLOAD_TOO_LARGE",
			Message:       "body excede o teto",
			CorrelationID: correlationID,
		})
	}
	if h.billing == nil {
		return c.JSON(http.StatusBadGateway, pkg.ErrorResponse{
			Error:         "GATEWAY_UNAVAILABLE",
			Message:       "Billing indisponível",
			CorrelationID: correlationID,
		})
	}
	result, status, callErr := h.billing.ForwardAsaasWebhook(c.Request().Context(), token, body)
	if callErr != nil {
		if httpErr, ok := callErr.(*pkg.AppError); ok {
			return c.JSON(httpErr.StatusCode, httpErr.WithTraceID(correlationID).ToResponse())
		}
		return handleError(c, callErr, correlationID)
	}
	if status == 0 {
		status = http.StatusOK
	}
	return writeRaw(c, status, result)
}

func (h *BillingHandlers) proxy(c echo.Context, status int, fn func(echo.Context, port.BillingScope) (json.RawMessage, error)) error {
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	result, err := fn(c, scope)
	if err != nil {
		return handleError(c, err, scope.CorrelationID)
	}
	return writeRaw(c, status, result)
}

func (h *BillingHandlers) guard(c echo.Context) (port.BillingScope, error) {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if h.billing == nil {
		return port.BillingScope{}, c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Billing indisponível",
			CorrelationID: correlationID,
		})
	}
	claims := middlewarePkg.GetClaimsFromContext(c)
	companyID := port.CompanyIDFromContext(c.Request().Context())
	if companyID == "" {
		return port.BillingScope{}, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_TENANT",
			Message:       "company do token JWT é obrigatória",
			CorrelationID: correlationID,
		})
	}
	userID := ""
	admin := false
	if claims != nil {
		userID = strings.TrimSpace(claims.UserID)
		if userID == "" {
			userID = strings.TrimSpace(claims.Sub)
		}
		admin = middlewarePkg.IsBillingOrgCaller(claims)
	}
	if userID == "" {
		return port.BillingScope{}, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "userId do token JWT é obrigatório",
			CorrelationID: correlationID,
		})
	}
	if token := middlewarePkg.GetTokenFromContext(c); token != "" {
		ctx := port.WithBearerToken(c.Request().Context(), token)
		c.SetRequest(c.Request().WithContext(ctx))
	}
	return port.BillingScope{
		CompanyID:     companyID,
		UserID:        userID,
		CorrelationID: correlationID,
		Admin:         admin,
	}, nil
}

func bindBillingJSON(c echo.Context) (map[string]any, error) {
	correlationID := middlewarePkg.GetCorrelationID(c)
	body := map[string]any{}
	if err := c.Bind(&body); err != nil {
		return nil, c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "BAD_REQUEST",
			Message:       "Payload inválido",
			CorrelationID: correlationID,
		})
	}
	if containsCardData(body) {
		return nil, c.JSON(http.StatusUnprocessableEntity, pkg.ErrorResponse{
			Error:         "PAYMENT_METHOD_UNSUPPORTED",
			Message:       "PAN/CVV não são aceitos",
			CorrelationID: correlationID,
		})
	}
	return body, nil
}

func (h *BillingHandlers) attachPayerProfile(c echo.Context, body map[string]any, scope port.BillingScope) error {
	delete(body, "payerName")
	delete(body, "payerEmail")
	delete(body, "payerCpfCnpj")
	if h.users == nil {
		return nil
	}
	codeUser := middlewarePkg.GetUserIDFromContext(c)
	user, err := h.users.GetUserByCodeUser(
		c.Request().Context(),
		codeUser,
		middlewarePkg.GetTokenFromContext(c),
		middlewarePkg.GetTenantId(c),
		scope.CorrelationID,
	)
	if err != nil {
		return handleError(c, err, scope.CorrelationID)
	}
	cpf := ""
	name := strings.TrimSpace(user.Email)
	if user.PersonProfile != nil {
		cpf = digitsOnly(user.PersonProfile.CPF)
		if fullName := strings.TrimSpace(user.PersonProfile.FullName); fullName != "" {
			name = fullName
		}
	}
	if len(cpf) != 11 && len(cpf) != 14 {
		return c.JSON(http.StatusUnprocessableEntity, pkg.ErrorResponse{
			Error:         "PAYER_DOCUMENT_MISSING",
			Message:       "Complete o CPF no perfil para assinar com PIX ou boleto.",
			CorrelationID: scope.CorrelationID,
		})
	}
	if name != "" {
		body["payerName"] = name
	}
	if email := strings.TrimSpace(user.Email); email != "" {
		body["payerEmail"] = email
	}
	body["payerCpfCnpj"] = cpf
	return nil
}

func digitsOnly(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func containsCardData(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
			if normalized == "cardnumber" || normalized == "cvv" || normalized == "pan" || normalized == "cvc" {
				return true
			}
			if containsCardData(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsCardData(nested) {
				return true
			}
		}
	}
	return false
}
