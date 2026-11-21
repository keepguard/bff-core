package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/clientip"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/keepguard/bff-core/internal/adapters/in/http/dto"
	"github.com/keepguard/bff-core/internal/adapters/in/http/mapper"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/in/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/register"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterHandlers implementa os handlers HTTP para registro de usuários e consentimentos
type RegisterHandlers struct {
	registerInitUseCase    register.RegisterInitUseCase
	registerConfirmUseCase register.RegisterConfirmUseCase
	registerResendUseCase  register.RegisterResendUseCase
	logger                 *zap.Logger
}

// NewRegisterHandlers cria uma nova instância dos RegisterHandlers
func NewRegisterHandlers(
	registerInitUseCase register.RegisterInitUseCase,
	registerConfirmUseCase register.RegisterConfirmUseCase,
	registerResendUseCase register.RegisterResendUseCase,
	logger *zap.Logger,
) *RegisterHandlers {
	return &RegisterHandlers{
		registerInitUseCase:    registerInitUseCase,
		registerConfirmUseCase: registerConfirmUseCase,
		registerResendUseCase:  registerResendUseCase,
		logger:                 logger,
	}
}

// InitRegisterHandler trata requisições de inicialização de registro de usuário
// @Summary Inicializar registro de usuário
// @Description Inicia o processo de registro de um novo usuário no sistema (endpoint público - não requer token)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.RegisterInitRequestDTO true "Dados para inicialização do registro"
// @Success 201 {object} dto.RegisterInitResponseDTO "Registro inicializado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 409 {object} pkg.ErrorResponse "Email já cadastrado ou sessão ativa existe"
// @Failure 429 {object} pkg.ErrorResponse "Muitas tentativas (Rate limit excedido)"
// @Failure 503 {object} pkg.ErrorResponse "Serviço temporariamente indisponível (Circuit breaker)"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /api/v1/register/init [post]
func (h *RegisterHandlers) InitRegisterHandler(c echo.Context) error {
	// ========================================================================
	// EXTRAÇÃO DE HEADERS OBRIGATÓRIOS
	// ========================================================================
	correlationID := middlewarePkg.GetCorrelationID(c)

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente",
			zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       "Header X-Tenant-Id é obrigatório",
			CorrelationID: correlationID,
		})
	}

	// ========================================================================
	// BIND E VALIDAÇÃO DO BODY
	// ========================================================================
	var req dto.RegisterInitRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de inicialização de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	// ========================================================================
	// CRIAR COMANDO DE DOMÍNIO
	// ========================================================================
	// Definir valores padrão para campos opcionais
	acceptedMarketing := false
	if req.AcceptedMarketing != nil {
		acceptedMarketing = *req.AcceptedMarketing
	}

	command := appdto.NewRegisterInitCommand(
		req.NameFull,
		req.Email,
		req.Password,
		req.ConfirmPassword,
		req.Phone,
		req.HasAcceptedTermsAndPrivacy,
		acceptedMarketing,
		ipDoCadastro(c.Request(), req.IPAddress),
		req.UserAgent,
		req.Geolocation,
		req.Type,
		tenantId,
		correlationID,
	)
	command.CountryCode = req.CountryCode
	command.PreferredLocale = req.PreferredLocale

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de inicialização de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "VALIDATION_ERROR",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	// ========================================================================
	// EXECUTAR CASO DE USO
	// ========================================================================
	response, err := h.registerInitUseCase.Execute(c.Request().Context(), command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusCreated, mapper.ToRegisterInitResponse(response))
}

// ConfirmRegisterHandler trata requisições de confirmação de registro de usuário
// @Summary Confirmar registro de usuário
// @Description Confirma o registro de um novo usuário validando o token de verificação (endpoint público - não requer token)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.RegisterConfirmRequestDTO true "Dados para confirmação do registro"
// @Success 200 {object} dto.RegisterConfirmResponseDTO "Registro confirmado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Token inválido ou dados inválidos"
// @Failure 404 {object} pkg.ErrorResponse "Sessão não encontrada ou expirada"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /api/v1/register/confirm [post]
func (h *RegisterHandlers) ConfirmRegisterHandler(c echo.Context) error {
	// ========================================================================
	// EXTRAÇÃO DE HEADERS OBRIGATÓRIOS
	// ========================================================================
	correlationID := middlewarePkg.GetCorrelationID(c)

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente",
			zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       "Header X-Tenant-Id é obrigatório",
			CorrelationID: correlationID,
		})
	}

	clientId := c.Request().Header.Get("X-Client-ID")
	if clientId == "" {
		clientId = "keepguard-default-client"
	}

	// ========================================================================
	// BIND E VALIDAÇÃO DO BODY
	// ========================================================================
	var req dto.RegisterConfirmRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de confirmação de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	// ========================================================================
	// CRIAR COMANDO DE DOMÍNIO
	// ========================================================================
	command := appdto.NewMultiChannelRegisterConfirmCommand(
		req.Email,
		req.RegistrationSessionID,
		req.Token,
		req.EmailToken,
		req.SmsToken,
		req.WhatsAppToken,
		tenantId,
		correlationID,
		clientId,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de confirmação de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "VALIDATION_ERROR",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	// ========================================================================
	// EXECUTAR CASO DE USO
	// ========================================================================
	ctx := port.WithDeviceID(c.Request().Context(), strings.TrimSpace(c.Request().Header.Get("X-Device-Id")))
	response, err := h.registerConfirmUseCase.Execute(ctx, command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, mapper.ToRegisterConfirmResponse(response))
}

// ResendRegisterTokenHandler trata requisições de reenvio de token
// @Summary Reenviar token de registro
// @Description Reenvia o token de verificação para o email (endpoint público)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação"
// @Param X-Tenant-Id header string true "ID da aplicação (UUID)"
// @Param request body dto.RegisterResendRequestDTO true "Dados para reenvio"
// @Success 200 {object} dto.RegisterResendResponseDTO "Token reenviado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Dados inválidos"
// @Failure 404 {object} pkg.ErrorResponse "Sessão não encontrada"
// @Router /api/v1/register/resend [post]
func (h *RegisterHandlers) ResendRegisterTokenHandler(c echo.Context) error {
	// Extração de headers
	correlationID := middlewarePkg.GetCorrelationID(c)

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       "Header X-Tenant-Id é obrigatório",
			CorrelationID: correlationID,
		})
	}

	// Bind do body
	var req dto.RegisterResendRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de reenvio",
			zap.String("correlationId", correlationID),
			zap.Error(err))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	// Criar comando
	command := appdto.NewRegisterResendCommand(
		req.Email,
		req.RegistrationSessionID,
		tenantId,
		correlationID,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "VALIDATION_ERROR",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	// Executar use case
	response, err := h.registerResendUseCase.Execute(c.Request().Context(), command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, mapper.ToRegisterResendResponse(response))
}

// handleError trata erros de forma padronizada
func handleError(c echo.Context, err error, correlationID string) error {
	if err != nil && (strings.Contains(err.Error(), "circuit breaker is open") || strings.Contains(err.Error(), "circuit breaker is half-open")) {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_TEMPORARILY_UNAVAILABLE",
			Message:       "O serviço está temporariamente indisponível. Por favor, tente novamente em instantes.",
			CorrelationID: correlationID,
		})
	}

	// Trata erros da aplicação (AppError)
	if appErr, ok := err.(*pkg.AppError); ok {
		return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
	}

	// Trata erros HTTP (HTTPError) - MapHTTPError já extraiu a mensagem detalhada
	var httpErr *appdto.HTTPError
	if errors.As(err, &httpErr) {
		if retryAfter, ok := registrationInProgress(httpErr); ok {
			appErr := pkg.NewAppError("REGISTRATION_IN_PROGRESS", registrationInProgressMessage(retryAfter), http.StatusConflict)
			appErr.RetryAfter = retryAfter
			if retryAfter > 0 {
				c.Response().Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
			}
			return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
		}
		code := "HTTP_ERROR"
		if upstream := upstreamErrorCode(httpErr.Details); upstream != "" {
			code = upstream
		}
		appErr := pkg.NewAppError(code, httpErr.Message, httpErr.Code)
		return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
	}

	// Erro genérico
	return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
		Error:         "INTERNAL_ERROR",
		Message:       "Erro interno do servidor",
		CorrelationID: correlationID,
	})
}

// registrationInProgress detects the ms-auth 409 REGISTRATION_IN_PROGRESS (a PENDING_ACTIVATION account still inside
// its window, left by a failed registration that the reaper will clean). ms-auth sends a ProblemDetail with
// errorCode/retryAfter nested in "properties"; root-level fields are accepted too. Nothing else from the upstream
// body is forwarded.
func registrationInProgress(httpErr *appdto.HTTPError) (int64, bool) {
	if httpErr == nil || httpErr.Code != http.StatusConflict {
		return 0, false
	}
	details := strings.TrimSpace(httpErr.Details)
	if details == "" || details[0] != '{' {
		return 0, false
	}
	type fields struct {
		ErrorCode  string   `json:"errorCode"`
		Error      string   `json:"error"`
		RetryAfter *float64 `json:"retryAfter"`
	}
	var payload struct {
		fields
		Properties fields `json:"properties"`
	}
	if err := json.Unmarshal([]byte(details), &payload); err != nil {
		return 0, false
	}
	for _, f := range []fields{payload.Properties, payload.fields} {
		if f.ErrorCode == "REGISTRATION_IN_PROGRESS" || f.Error == "REGISTRATION_IN_PROGRESS" {
			var retryAfter int64
			if f.RetryAfter != nil && *f.RetryAfter > 0 {
				retryAfter = int64(math.Ceil(*f.RetryAfter))
			} else if payload.Properties.RetryAfter != nil && *payload.Properties.RetryAfter > 0 {
				retryAfter = int64(math.Ceil(*payload.Properties.RetryAfter))
			}
			return retryAfter, true
		}
	}
	return 0, false
}

// registrationInProgressMessage is the single pt-BR text for REGISTRATION_IN_PROGRESS shown to the person.
func registrationInProgressMessage(retryAfter int64) string {
	if retryAfter <= 0 {
		return "Já existe um cadastro em andamento com estes dados. Tente novamente em alguns minutos."
	}
	minutes := (retryAfter + 59) / 60
	if minutes == 1 {
		return "Já existe um cadastro em andamento com estes dados. Tente novamente em 1 minuto."
	}
	return fmt.Sprintf("Já existe um cadastro em andamento com estes dados. Tente novamente em %d minutos.", minutes)
}

func upstreamErrorCode(details string) string {
	details = strings.TrimSpace(details)
	if details == "" || details[0] != '{' {
		return ""
	}
	var payload struct {
		Error     string `json:"error"`
		ErrorCode string `json:"errorCode"`
	}
	if err := json.Unmarshal([]byte(details), &payload); err != nil {
		return ""
	}
	if code := strings.TrimSpace(payload.ErrorCode); code != "" && !strings.Contains(code, " ") {
		return code
	}
	code := strings.TrimSpace(payload.Error)
	if code == "" || strings.Contains(code, " ") {
		return ""
	}
	return code
}

// ipDoCadastro: o IP do antifraude do cadastro vem do pedido (X-Real-IP do nginx do front / borda), não do corpo, que o
// navegador escolhe. Antes ninguém preenchia o campo e o IP ficava vazio. Corpo só vale se o servidor não souber o IP.
func ipDoCadastro(r *http.Request, doCorpo string) string {
	if ip := clientip.FromRequest(r); ip != "" {
		return ip
	}
	return strings.TrimSpace(doCorpo)
}
