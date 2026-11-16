package register

import (
	"context"
	"errors"
	"strings"
	"testing"

	companyDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/company"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const initTestEmail = "novo.usuario@example.com"

// newInitFixture monta o init com o caminho feliz; o publisher falha quando failures > 0 (vezes que falha antes).
func newInitUC(t *testing.T, pub *MockMessagePublisher, opts RegisterOptions, logger *zap.Logger) *registerInitUseCaseImpl {
	t.Helper()
	company := companyDto.MSCompanyResponseDTO{ID: "company-123", Name: "Test Company"}
	companyClient := &MockCompanyClient{}
	companyClient.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(company, nil)

	userClient := &MockUserClient{}
	userClient.On("InitRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(
		appdto.MSUserRegisterInitResponseDTO{
			RegistrationSessionID: "3f1c2a9e-0000-4000-8000-000000000001",
			Email:                 initTestEmail,
			ExpiresIn:             900,
			Token:                 "tok-secret-do-init",
		}, nil)

	return &registerInitUseCaseImpl{
		authClient:          &MockAuthClient{},
		userClient:          userClient,
		companyClient:       companyClient,
		communicationClient: &MockCommunicationClient{},
		messagePublisher:    pub,
		logger:              logger,
		opts:                opts,
	}
}

func initCommand() appdto.RegisterInitCommand {
	return appdto.RegisterInitCommand{
		NameFull:                   "Novo Usuario",
		Email:                      initTestEmail,
		Password:                   "senha-forte-123",
		ConfirmPassword:            "senha-forte-123",
		Phone:                      "+5511999999999",
		HasAcceptedTermsAndPrivacy: true,
		Type:                       "PERSON",
		TenantId:                   "test-app",
		CorrelationID:              "corr-123",
	}
}

// Flag desligada: uma tentativa só (como antes), init responde normalmente e não há campo resendAvailable.
func TestInitSend_FlagOffSingleAttemptNoResendField(t *testing.T) {
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(errors.New("broker fora"))
	uc := newInitUC(t, pub, RegisterOptions{SagaHardening: false}, zap.NewNop())

	view, err := uc.Execute(context.Background(), initCommand())

	require.NoError(t, err)
	assert.Nil(t, view.ResendAvailable)
	pub.AssertNumberOfCalls(t, "PublishMessage", 1)
}

// Flag ligada: falha transitória do envio é retentada e o init indica que o reenvio está disponível.
func TestInitSend_FlagOnRetriesThenSucceeds(t *testing.T) {
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(errors.New("broker piscou")).Twice()
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)
	metrics := newFakeRegistrationMetrics()
	uc := newInitUC(t, pub, RegisterOptions{SagaHardening: true, Metrics: metrics}, zap.NewNop())

	view, err := uc.Execute(context.Background(), initCommand())

	require.NoError(t, err)
	require.NotNil(t, view.ResendAvailable)
	assert.True(t, *view.ResendAvailable)
	pub.AssertNumberOfCalls(t, "PublishMessage", 3)
	assert.Equal(t, 2, metrics.step("InitSendCode", "error"))
	assert.Equal(t, 1, metrics.step("InitSendCode", "ok"))
}

// Flag ligada e envio sempre falha: 3 tentativas, o init ainda responde (sem erro) e conta as falhas.
func TestInitSend_FlagOnAllAttemptsFailInitStillSucceeds(t *testing.T) {
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(errors.New("broker fora"))
	metrics := newFakeRegistrationMetrics()
	uc := newInitUC(t, pub, RegisterOptions{SagaHardening: true, Metrics: metrics}, zap.NewNop())

	view, err := uc.Execute(context.Background(), initCommand())

	require.NoError(t, err)
	require.NotNil(t, view.ResendAvailable)
	pub.AssertNumberOfCalls(t, "PublishMessage", 3)
	assert.Equal(t, 3, metrics.step("InitSendCode", "error"))
}

// O log da falha de envio não leva e-mail, código nem token.
func TestInitSend_LogHasNoPersonalData(t *testing.T) {
	core, recorded := observer.New(zap.WarnLevel)
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(errors.New("falha de transporte"))
	uc := newInitUC(t, pub, RegisterOptions{SagaHardening: true, Metrics: newFakeRegistrationMetrics()}, zap.New(core))

	_, err := uc.Execute(context.Background(), initCommand())
	require.NoError(t, err)

	entries := recorded.All()
	require.NotEmpty(t, entries, "a falha deveria ser logada")
	for _, e := range entries {
		dump := e.Message
		for _, f := range e.Context {
			dump += " " + f.Key + "=" + f.String
			if f.Interface != nil {
				dump += " " + strings.ReplaceAll(strings.TrimSpace(anyString(f.Interface)), "\n", " ")
			}
		}
		assert.NotContains(t, dump, initTestEmail)
		assert.NotContains(t, dump, "tok-secret-do-init")
		assert.NotContains(t, dump, "senha-forte-123")
	}
}

func anyString(v interface{}) string {
	if s, ok := v.(error); ok {
		return s.Error()
	}
	return ""
}

// O publisher recebe a mensagem original (o retry não altera o conteúdo).
func TestInitSend_RetrySendsSameMessage(t *testing.T) {
	var seen []outport.MessageDTO
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		seen = append(seen, a.Get(1).(outport.MessageDTO))
	}).Return(errors.New("x")).Once()
	pub.On("PublishMessage", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		seen = append(seen, a.Get(1).(outport.MessageDTO))
	}).Return(nil)
	uc := newInitUC(t, pub, RegisterOptions{SagaHardening: true, Metrics: newFakeRegistrationMetrics()}, zap.NewNop())

	_, err := uc.Execute(context.Background(), initCommand())
	require.NoError(t, err)

	require.Len(t, seen, 2)
	assert.Equal(t, seen[0], seen[1])
}
