package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-core/internal/application/billing"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
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

func (h *BillingHandlers) ListBillingGatewayAccountsHandler(c echo.Context) error {
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.ListGatewayAccounts(ctx.Request().Context(), scope)
	})
}

func (h *BillingHandlers) PutBillingGatewayAccountByGatewayHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	gateway := c.Param("gateway")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.PutGatewayAccountByGateway(ctx.Request().Context(), scope, gateway, body)
	})
}

func (h *BillingHandlers) SetPrimaryBillingGatewayHandler(c echo.Context) error {
	gateway := c.Param("gateway")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.SetPrimaryGateway(ctx.Request().Context(), scope, gateway)
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
	if c.Response().Committed {
		return nil
	}
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	claims := middlewarePkg.GetClaimsFromContext(c)
	// MANAGER opera a org e não assina. ADMIN/SYSTEM podem ter plano próprio.
	if claims != nil && pkg.HasAnyRole(claims.Roles, "MANAGER") && !pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") {
		return handleError(c, pkg.NewAppError(
			"BILLING_PAYER_OPS",
			"Quem opera a organização não assina",
			http.StatusForbidden,
		), scope.CorrelationID)
	}
	if attachErr := h.attachPayerProfile(c, body, scope); attachErr != nil {
		return handleError(c, attachErr, scope.CorrelationID)
	}
	h.attachSessionSnapshot(c, body, claims)
	// Assinatura é sempre do caller: não marcar X-Caller-Admin (evita BILLING_PAYER_OPS no MS).
	payerScope := scope
	payerScope.Admin = false
	result, status, callErr := h.billing.CreateSubscription(c.Request().Context(), payerScope, body)
	if callErr != nil {
		return handleError(c, callErr, scope.CorrelationID)
	}
	if status == 0 {
		status = http.StatusCreated
	}
	return writeRaw(c, status, result)
}

func (h *BillingHandlers) GrantLifetimeBillingSubscriptionHandler(c echo.Context) error {
	body, err := bindBillingJSON(c)
	if err != nil {
		return err
	}
	return h.proxy(c, http.StatusCreated, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.GrantLifetimeSubscription(ctx.Request().Context(), scope, body)
	})
}

func (h *BillingHandlers) LookupBillingUserHandler(c echo.Context) error {
	scope, unavailable := h.guard(c)
	if unavailable != nil {
		return unavailable
	}
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		q = strings.TrimSpace(c.QueryParam("email"))
	}
	if q == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "VALIDATION_ERROR",
			Message:       "Informe o e-mail do usuário para busca",
			CorrelationID: scope.CorrelationID,
		})
	}
	if h.users == nil {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_UNAVAILABLE",
			Message:       "Serviço de usuários indisponível",
			CorrelationID: scope.CorrelationID,
		})
	}
	tenantID := middlewarePkg.GetTenantId(c)
	user, err := h.users.GetByEmail(c.Request().Context(), q, tenantID, scope.CompanyID, scope.CorrelationID)
	if err != nil {
		h.logger.Warn("Falha ao buscar usuário por email no lookup de billing",
			zap.String("correlationId", scope.CorrelationID),
			zap.String("companyId", scope.CompanyID),
			zap.String("email", q),
			zap.Error(err),
		)
		return c.JSON(http.StatusNotFound, pkg.ErrorResponse{
			Error:         "USER_NOT_FOUND",
			Message:       "Usuário não encontrado nesta organização",
			CorrelationID: scope.CorrelationID,
		})
	}
	userID := strings.TrimSpace(user.CodeUser)
	if userID == "" {
		userID = strings.TrimSpace(user.ID)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":            userID,
		"codeUser":      user.CodeUser,
		"username":      user.Username,
		"email":         user.Email,
		"status":        user.Status,
		"emailVerified": user.EmailVerified,
	})
}

func (h *BillingHandlers) CancelBillingSubscriptionHandler(c echo.Context) error {
	id := c.Param("id")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		return h.billing.CancelSubscription(ctx.Request().Context(), scope, id)
	})
}

func (h *BillingHandlers) ListBillingInvoicesHandler(c echo.Context) error {
	query := billingListQuery(c, "page", "size", "status", "paymentMethod", "payerUserId", "from", "to", "q")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		resolved := h.resolvePayerQuery(ctx, scope, query)
		raw, err := h.billing.ListInvoices(ctx.Request().Context(), scope, resolved)
		if err != nil {
			return nil, err
		}
		if !scope.Admin {
			return raw, nil
		}
		return h.enrichBillingPage(ctx, scope, raw, "payerUserId"), nil
	})
}

func (h *BillingHandlers) ListBillingEntitlementsHandler(c echo.Context) error {
	query := billingListQuery(c, "page", "size", "status", "planCode", "userId", "q", "hasPlan")
	return h.proxy(c, http.StatusOK, func(ctx echo.Context, scope port.BillingScope) (json.RawMessage, error) {
		if !scope.Admin {
			return nil, pkg.NewAppError("FORBIDDEN", "Somente quem opera a organização lista entitlements.", http.StatusForbidden)
		}
		resolved := h.resolvePayerQuery(ctx, scope, query)
		raw, err := h.billing.ListEntitlements(ctx.Request().Context(), scope, resolved)
		if err != nil {
			return nil, err
		}
		return h.enrichBillingPage(ctx, scope, raw, "userId"), nil
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

func (h *BillingHandlers) StripeWebhookHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if c.QueryString() != "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "query string recusada no webhook",
			CorrelationID: correlationID,
		})
	}
	token := strings.TrimSpace(c.Request().Header.Get("stripe-signature"))
	if token == "" {
		token = strings.TrimSpace(c.Request().Header.Get("x-webhook-token"))
	}
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "webhook token/signature inválido",
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
	result, status, callErr := h.billing.ForwardStripeWebhook(c.Request().Context(), token, body)
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
	method := strings.ToLower(strings.TrimSpace(stringifyJSON(body["paymentMethod"])))
	if method == "credit_card" || method == "creditcard" {
		token := strings.TrimSpace(stringifyJSON(body["creditCardToken"]))
		if token == "" {
			return nil, c.JSON(http.StatusUnprocessableEntity, pkg.ErrorResponse{
				Error:         "CARD_TOKEN_MISSING",
				Message:       "Informe o token do cartão (Asaas)",
				CorrelationID: correlationID,
			})
		}
		body["paymentMethod"] = "credit_card"
		body["creditCardToken"] = token
		if remote := clientRemoteIP(c); remote != "" {
			body["remoteIp"] = remote
		}
	}
	return body, nil
}

func clientRemoteIP(c echo.Context) string {
	if xff := strings.TrimSpace(c.Request().Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(c.Request().Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return strings.TrimSpace(c.Request().RemoteAddr)
	}
	return host
}

func (h *BillingHandlers) attachPayerProfile(c echo.Context, body map[string]any, scope port.BillingScope) error {
	if body == nil {
		return nil
	}
	clientDoc := digitsOnly(stringifyJSON(body["payerCpfCnpj"]))
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
		return err
	}
	name := strings.TrimSpace(user.Email)
	cpf := ""
	if user.PersonProfile != nil {
		cpf = digitsOnly(user.PersonProfile.CPF)
		if fullName := strings.TrimSpace(user.PersonProfile.FullName); fullName != "" {
			name = fullName
		}
	}
	if len(cpf) != 11 && len(cpf) != 14 {
		if err := h.firstWritePayerDocument(c, &user, clientDoc, scope); err != nil {
			return err
		}
		cpf = ""
		if user.PersonProfile != nil {
			cpf = digitsOnly(user.PersonProfile.CPF)
			if fullName := strings.TrimSpace(user.PersonProfile.FullName); fullName != "" {
				name = fullName
			}
		}
	}
	if name != "" {
		body["payerName"] = name
	}
	if email := strings.TrimSpace(user.Email); email != "" {
		body["payerEmail"] = email
	}
	if len(cpf) == 11 || len(cpf) == 14 {
		body["payerCpfCnpj"] = cpf
	}
	return nil
}

func (h *BillingHandlers) attachSessionSnapshot(c echo.Context, body map[string]any, claims *pkg.JWTClaims) {
	if body == nil {
		return
	}
	if _, exists := body["deviceId"]; !exists || stringifyJSON(body["deviceId"]) == "" {
		devID := strings.TrimSpace(c.Request().Header.Get("X-Device-Id"))
		if devID == "" && claims != nil {
			devID = strings.TrimSpace(claims.DeviceID)
		}
		if devID != "" {
			body["deviceId"] = devID
		}
	}

	if _, exists := body["deviceName"]; !exists || stringifyJSON(body["deviceName"]) == "" {
		if devName := strings.TrimSpace(c.Request().Header.Get("X-Device-Name")); devName != "" {
			body["deviceName"] = devName
		}
	}

	if _, exists := body["deviceType"]; !exists || stringifyJSON(body["deviceType"]) == "" {
		if devType := strings.TrimSpace(c.Request().Header.Get("X-Device-Type")); devType != "" {
			body["deviceType"] = devType
		}
	}

	if _, exists := body["ipAddress"]; !exists || stringifyJSON(body["ipAddress"]) == "" {
		ip := strings.TrimSpace(c.Request().Header.Get("X-Public-IP"))
		if ip == "" {
			ip = strings.TrimSpace(c.RealIP())
		}
		if ip != "" {
			body["ipAddress"] = ip
		}
	}

	if _, exists := body["userAgent"]; !exists || stringifyJSON(body["userAgent"]) == "" {
		if ua := strings.TrimSpace(c.Request().UserAgent()); ua != "" {
			body["userAgent"] = ua
		}
	}

	if _, exists := body["sessionId"]; !exists || stringifyJSON(body["sessionId"]) == "" {
		sessID := strings.TrimSpace(c.Request().Header.Get("X-Session-Id"))
		if sessID == "" && claims != nil {
			sessID = strings.TrimSpace(claims.JTI)
		}
		if sessID == "" && claims != nil {
			sessID = strings.TrimSpace(claims.DeviceID)
		}
		if sessID != "" {
			body["sessionId"] = sessID
		}
	}
}

func (h *BillingHandlers) firstWritePayerDocument(c echo.Context, user *appdto.MSUserResponseDTO, clientDoc string, scope port.BillingScope) error {
	if len(clientDoc) != 11 {
		return pkg.NewAppError("PAYER_DOCUMENT_MISSING", "Informe um CPF válido.", http.StatusUnprocessableEntity)
	}
	if !brazilianCPFValid(clientDoc) {
		return pkg.NewAppError("PAYER_DOCUMENT_INVALID", "Informe um CPF válido.", http.StatusUnprocessableEntity)
	}
	userID := strings.TrimSpace(user.ID)
	if userID == "" {
		return pkg.NewAppError("PAYER_DOCUMENT_MISSING", "Informe um CPF válido.", http.StatusUnprocessableEntity)
	}
	patched, err := h.users.PatchPersonDocument(
		c.Request().Context(),
		userID,
		clientDoc,
		middlewarePkg.GetTokenFromContext(c),
		middlewarePkg.GetTenantId(c),
		scope.CorrelationID,
	)
	if err != nil {
		return mapPersonDocumentError(err)
	}
	reloaded, reloadErr := h.users.GetUserByCodeUser(
		c.Request().Context(),
		middlewarePkg.GetUserIDFromContext(c),
		middlewarePkg.GetTokenFromContext(c),
		middlewarePkg.GetTenantId(c),
		scope.CorrelationID,
	)
	if reloadErr == nil {
		*user = reloaded
	} else {
		*user = patched
	}
	if user.PersonProfile == nil {
		user.PersonProfile = &appdto.PersonProfileDTO{}
	}
	if len(digitsOnly(user.PersonProfile.CPF)) != 11 {
		user.PersonProfile.CPF = clientDoc
	}
	return nil
}

func mapPersonDocumentError(err error) error {
	httpErr, ok := err.(*appdto.HTTPError)
	if !ok {
		return err
	}
	code := upstreamErrorCode(httpErr.Details)
	if code == "CPF_ALREADY_EXISTS" || (httpErr.Code == http.StatusConflict && strings.Contains(strings.ToUpper(httpErr.Details+" "+httpErr.Message), "CPF_ALREADY_EXISTS")) {
		return pkg.NewAppError("CPF_ALREADY_EXISTS", "Este CPF já está em uso nesta organização.", http.StatusConflict)
	}
	if code == "PAYER_DOCUMENT_INVALID" || httpErr.Code == http.StatusUnprocessableEntity {
		return pkg.NewAppError("PAYER_DOCUMENT_INVALID", "Informe um CPF válido.", http.StatusUnprocessableEntity)
	}
	if code == "PAYER_DOCUMENT_IMMUTABLE" {
		return pkg.NewAppError("PAYER_DOCUMENT_IMMUTABLE", "Documento do pagador já está cadastrado", http.StatusConflict)
	}
	return err
}

func brazilianCPFValid(digits string) bool {
	if len(digits) != 11 {
		return false
	}
	allSame := true
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	rest := sum * 10 % 11
	if rest == 10 {
		rest = 0
	}
	if rest != int(digits[9]-'0') {
		return false
	}
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	rest = sum * 10 % 11
	if rest == 10 {
		rest = 0
	}
	return rest == int(digits[10]-'0')
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

func billingListQuery(c echo.Context, keys ...string) map[string]string {
	query := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.QueryParam(key)); value != "" {
			query[key] = value
		}
	}
	return query
}

func (h *BillingHandlers) resolvePayerQuery(c echo.Context, scope port.BillingScope, query map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range query {
		out[key] = value
	}
	q := strings.TrimSpace(out["q"])
	if q == "" {
		return out
	}
	if looksLikeUUID(q) {
		if out["userId"] == "" {
			out["userId"] = q
		}
		if out["payerUserId"] == "" {
			out["payerUserId"] = q
		}
		return out
	}
	if strings.Contains(q, "@") && h.users != nil {
		user, err := h.users.GetByEmail(
			c.Request().Context(),
			q,
			middlewarePkg.GetTenantId(c),
			scope.CompanyID,
			scope.CorrelationID,
		)
		if err == nil {
			id := strings.TrimSpace(user.CodeUser)
			if id == "" {
				id = strings.TrimSpace(user.ID)
			}
			if id != "" {
				out["userId"] = id
				out["payerUserId"] = id
				delete(out, "q")
				return out
			}
		}
		out["userId"] = "00000000-0000-0000-0000-000000000000"
		out["payerUserId"] = "00000000-0000-0000-0000-000000000000"
		delete(out, "q")
	}
	return out
}

func (h *BillingHandlers) enrichBillingPage(c echo.Context, scope port.BillingScope, raw json.RawMessage, idField string) json.RawMessage {
	if h.users == nil || len(raw) == 0 {
		return raw
	}
	var page map[string]any
	if err := json.Unmarshal(raw, &page); err != nil {
		return raw
	}
	items, _ := page["items"].([]any)
	if len(items) == 0 {
		return raw
	}
	cache := map[string]payerProfile{}
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := stringifyJSON(row[idField])
		if id == "" {
			continue
		}
		profile, found := cache[id]
		if !found {
			profile = h.lookupPayer(c, scope, id)
			cache[id] = profile
		}
		if profile.name != "" {
			row["payerName"] = profile.name
		}
		if profile.email != "" {
			row["payerEmail"] = profile.email
		}
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		return raw
	}
	return encoded
}

type payerProfile struct {
	name  string
	email string
}

func (h *BillingHandlers) lookupPayer(c echo.Context, scope port.BillingScope, userID string) payerProfile {
	user, err := h.users.GetUserByCodeUser(
		c.Request().Context(),
		userID,
		middlewarePkg.GetTokenFromContext(c),
		middlewarePkg.GetTenantId(c),
		scope.CorrelationID,
	)
	if err != nil {
		return payerProfile{}
	}
	name := strings.TrimSpace(user.Email)
	if user.PersonProfile != nil {
		if fullName := strings.TrimSpace(user.PersonProfile.FullName); fullName != "" {
			name = fullName
		}
	}
	return payerProfile{name: name, email: strings.TrimSpace(user.Email)}
}

func looksLikeUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
			continue
		}
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func stringifyJSON(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
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
