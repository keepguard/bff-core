package handlers

import (
	"net/http"

	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/inbound/http/middleware"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/register"
	"github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RegisterHandlers implementa os handlers HTTP para registro de usuários e consentimentos
type RegisterHandlers struct {
	registerInitUseCase    register.RegisterInitUseCase
	registerConfirmUseCase register.RegisterConfirmUseCase
	registerResendUseCase  register.RegisterResendUseCase
	consentDocumentClient  client.ConsentDocumentClient
	logger                 *zap.Logger
}

// NewRegisterHandlers cria uma nova instância dos RegisterHandlers
func NewRegisterHandlers(
	registerInitUseCase register.RegisterInitUseCase,
	registerConfirmUseCase register.RegisterConfirmUseCase,
	registerResendUseCase register.RegisterResendUseCase,
	consentDocumentClient client.ConsentDocumentClient,
	logger *zap.Logger,
) *RegisterHandlers {
	return &RegisterHandlers{
		registerInitUseCase:    registerInitUseCase,
		registerConfirmUseCase: registerConfirmUseCase,
		registerResendUseCase:  registerResendUseCase,
		consentDocumentClient:  consentDocumentClient,
		logger:                 logger,
	}
}

// InitRegisterHandler trata requisições de inicialização de registro de usuário
// @Summary Inicializar registro de usuário
// @Description Inicia o processo de registro de um novo usuário no sistema (endpoint público - não requer token)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.RegisterInitRequestDTO true "Dados para inicialização do registro"
// @Success 201 {object} dto.RegisterInitResponseDTO "Registro inicializado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 409 {object} pkg.ErrorResponse "Email já cadastrado ou sessão ativa existe"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /api/v1/register/init [post]
func (h *RegisterHandlers) InitRegisterHandler(c echo.Context) error {
	// ========================================================================
	// EXTRAÇÃO DE HEADERS OBRIGATÓRIOS
	// ========================================================================
	correlationID := middlewarePkg.GetCorrelationID(c)
	if correlationID == "" {
		h.logger.Warn("X-Correlation-ID ausente")
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Correlation-ID é obrigatório",
			TraceID: middlewarePkg.GetTraceID(c),
		})
	}

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente",
			zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Tenant-Id é obrigatório",
			TraceID: correlationID,
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
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
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
		req.IPAddress,
		req.UserAgent,
		req.Geolocation,
		req.Type,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de inicialização de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// ========================================================================
	// EXECUTAR CASO DE USO
	// ========================================================================
	response, err := h.registerInitUseCase.Execute(command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusCreated, response)
}

// ConfirmRegisterHandler trata requisições de confirmação de registro de usuário
// @Summary Confirmar registro de usuário
// @Description Confirma o registro de um novo usuário validando o token de verificação (endpoint público - não requer token)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
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
	if correlationID == "" {
		h.logger.Warn("X-Correlation-ID ausente")
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Correlation-ID é obrigatório",
			TraceID: middlewarePkg.GetTraceID(c),
		})
	}

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente",
			zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Tenant-Id é obrigatório",
			TraceID: correlationID,
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
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
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
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de confirmação de registro",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// ========================================================================
	// EXECUTAR CASO DE USO
	// ========================================================================
	response, err := h.registerConfirmUseCase.Execute(command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, response)
}

// ResendRegisterTokenHandler trata requisições de reenvio de token
// @Summary Reenviar token de registro
// @Description Reenvia o token de verificação para o email (endpoint público)
// @Tags register
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação"
// @Param X-Tenant-Id header string true "ID da aplicação (UUID)"
// @Param request body dto.RegisterResendRequestDTO true "Dados para reenvio"
// @Success 200 {object} dto.RegisterResendResponseDTO "Token reenviado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Dados inválidos"
// @Failure 404 {object} pkg.ErrorResponse "Sessão não encontrada"
// @Router /api/v1/register/resend [post]
func (h *RegisterHandlers) ResendRegisterTokenHandler(c echo.Context) error {
	// Extração de headers
	correlationID := middlewarePkg.GetCorrelationID(c)
	if correlationID == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Correlation-ID é obrigatório",
			TraceID: middlewarePkg.GetTraceID(c),
		})
	}

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Tenant-Id é obrigatório",
			TraceID: correlationID,
		})
	}

	// Bind do body
	var req dto.RegisterResendRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de reenvio",
			zap.String("correlationId", correlationID),
			zap.Error(err))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando
	command := appdto.NewRegisterResendCommand(
		req.Email,
		req.RegistrationSessionID,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar use case
	response, err := h.registerResendUseCase.Execute(command)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, response)
}

// GetPublishedConsentsHandler retorna todos os documentos de consentimento publicados
// @Summary Listar documentos de consentimento publicados
// @Description Retorna todos os termos e políticas publicados disponíveis para aceite (endpoint público)
// @Tags consents
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Success 200 {array} dto.ConsentDocumentResponseDTO "Lista de documentos de consentimento publicados"
// @Failure 400 {object} pkg.ErrorResponse "Headers obrigatórios ausentes"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /api/v1/consents/published [get]
func (h *RegisterHandlers) GetPublishedConsentsHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if correlationID == "" {
		h.logger.Warn("X-Correlation-ID ausente")
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Correlation-ID é obrigatório",
			TraceID: middlewarePkg.GetTraceID(c),
		})
	}

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Tenant-Id é obrigatório",
			TraceID: correlationID,
		})
	}

	docs, err := h.consentDocumentClient.FindAllPublished(c.Request().Context(), "", tenantId, correlationID)
	if err != nil {
		h.logger.Error("Erro ao buscar documentos publicados",
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
			Error:   "INTERNAL_SERVER_ERROR",
			Message: "Erro ao buscar documentos de consentimento",
			TraceID: correlationID,
		})
	}

	return c.JSON(http.StatusOK, docs)
}

// GetLatestByTypeHandler retorna a última versão publicada de um tipo específico de consentimento
// @Summary Buscar último documento publicado por tipo
// @Description Retorna o documento publicado mais recente de um tipo específico (ex: TERMS_OF_USE, PRIVACY_POLICY)
// @Tags consents
// @Produce json
// @Param type path string true "Tipo de consentimento (TERMS_OF_USE, PRIVACY_POLICY, LGPD_COMPLIANCE)"
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Success 200 {object} dto.ConsentDocumentResponseDTO "Documento de consentimento publicado"
// @Failure 400 {object} pkg.ErrorResponse "Headers obrigatórios ausentes"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /api/v1/consents/type/{type}/latest [get]
func (h *RegisterHandlers) GetLatestByTypeHandler(c echo.Context) error {
	correlationID := middlewarePkg.GetCorrelationID(c)
	if correlationID == "" {
		h.logger.Warn("X-Correlation-ID ausente")
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Correlation-ID é obrigatório",
			TraceID: middlewarePkg.GetTraceID(c),
		})
	}

	tenantId := middlewarePkg.GetTenantId(c)
	if tenantId == "" {
		h.logger.Warn("X-Tenant-Id ausente", zap.String("correlationId", correlationID))
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: "Header X-Tenant-Id é obrigatório",
			TraceID: correlationID,
		})
	}

	consentType := c.Param("type")
	if consentType == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_PARAM",
			Message: "Parâmetro type é obrigatório",
			TraceID: correlationID,
		})
	}

	doc, err := h.consentDocumentClient.FindLatestPublishedByType(c.Request().Context(), consentType, "", tenantId, correlationID)
	if err != nil {
		h.logger.Error("Erro ao buscar documento por tipo",
			zap.String("type", consentType),
			zap.String("correlationId", correlationID),
			zap.String("tenantId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
			Error:   "INTERNAL_SERVER_ERROR",
			Message: "Erro ao buscar documento de consentimento",
			TraceID: correlationID,
		})
	}

	return c.JSON(http.StatusOK, doc)
}

// handleError trata erros de forma padronizada
func handleError(c echo.Context, err error, correlationID string) error {
	// Trata erros da aplicação (AppError)
	if appErr, ok := err.(*pkg.AppError); ok {
		return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
	}

	// Trata erros HTTP (HTTPError) - MapHTTPError já extraiu a mensagem detalhada
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		appErr := pkg.NewAppError("HTTP_ERROR", httpErr.Message, httpErr.Code)
		return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
	}

	// Erro genérico
	return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "Erro interno do servidor",
		TraceID: correlationID,
	})
}
