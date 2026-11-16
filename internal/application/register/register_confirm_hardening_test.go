package register

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	authDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/auth"
	companyDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/company"
	userDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/user"
	userConsentDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/user_consent"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/domain/saga"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------------------------------------------
// Fakes de teste (só em memória de teste; em produção o estado vai para o Redis, ver redisstore).
// ---------------------------------------------------------------------------------------------------------------

type fakeRegisterState struct {
	mu           sync.Mutex
	locks        map[string]bool
	results      map[string][]byte
	lockErr      error
	tryLockCalls int
}

func newFakeRegisterState() *fakeRegisterState {
	return &fakeRegisterState{locks: map[string]bool{}, results: map[string][]byte{}}
}

func (f *fakeRegisterState) TryLock(_ context.Context, key string, _ time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tryLockCalls++
	if f.lockErr != nil {
		return false, f.lockErr
	}
	if f.locks[key] {
		return false, nil
	}
	f.locks[key] = true
	return true, nil
}

func (f *fakeRegisterState) Unlock(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.locks, key)
	return nil
}

func (f *fakeRegisterState) GetResult(_ context.Context, key string) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.results[key]
	return v, ok, nil
}

func (f *fakeRegisterState) SetResult(_ context.Context, key string, value []byte, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.results[key] = append([]byte(nil), value...)
	return nil
}

func (f *fakeRegisterState) lockHeld(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.locks[key]
}

func (f *fakeRegisterState) tryLocks() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tryLockCalls
}

type fakeRegistrationMetrics struct {
	mu    sync.Mutex
	steps map[string]int
	comps map[string]int
}

func newFakeRegistrationMetrics() *fakeRegistrationMetrics {
	return &fakeRegistrationMetrics{steps: map[string]int{}, comps: map[string]int{}}
}

func (m *fakeRegistrationMetrics) RecordStep(step, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps[step+"|"+result]++
}

func (m *fakeRegistrationMetrics) RecordCompensation(step, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.comps[step+"|"+result]++
}

func (m *fakeRegistrationMetrics) step(step, result string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.steps[step+"|"+result]
}

func (m *fakeRegistrationMetrics) comp(step, result string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.comps[step+"|"+result]
}

// ---------------------------------------------------------------------------------------------------------------
// Fixture: mocks com o caminho feliz; overrides rodam antes e ganham (testify usa a primeira expectativa).
// ---------------------------------------------------------------------------------------------------------------

type confirmFixture struct {
	user    *MockUserClient
	auth    *MockAuthClient
	consent *MockUserConsentClient
	company *MockCompanyClient
	comm    *MockCommunicationClient
	pub     *MockMessagePublisher
}

var (
	hfCompany = companyDto.MSCompanyResponseDTO{ID: "company-123", Name: "Test Company"}
	hfConfirm = userDto.MSUserRegisterConfirmResponseDTO{
		Email: "test@example.com", NameFull: "Test User", Phone: "+5511999999999", Type: "PERSON",
	}
	hfUser  = userDto.MSUserResponseDTO{ID: "user-123", Email: "test@example.com", CompanyID: "company-123", CodeUser: "code-123"}
	hfAuth  = authDto.AuthUserCreateResponseDTO{ID: "auth-123", CodeUser: "code-123", Email: "test@example.com", Status: "ACTIVE"}
	hfLogin = authDto.AuthLoginResponseDTO{Token: "jwt-token-123", ExpiresIn: 3600}
)

func testConfirmCommand() appdto.RegisterConfirmCommand {
	return appdto.RegisterConfirmCommand{
		Email:                 "test@example.com",
		RegistrationSessionId: "3f1c2a9e-0000-4000-8000-000000000001",
		Token:                 "123456",
		EmailToken:            "654321",
		TenantId:              "test-app",
		CorrelationID:         "corr-123",
		ClientId:              "client-1",
	}
}

func newConfirmFixture(overrides func(f *confirmFixture)) *confirmFixture {
	f := &confirmFixture{
		user: &MockUserClient{}, auth: &MockAuthClient{}, consent: &MockUserConsentClient{},
		company: &MockCompanyClient{}, comm: &MockCommunicationClient{}, pub: &MockMessagePublisher{},
	}
	if overrides != nil {
		overrides(f)
	}
	f.company.On("GetByTenantId", mock.Anything, "test-app", "corr-123").Return(hfCompany, nil)
	f.user.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(hfConfirm, nil)
	f.user.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(hfUser, nil)
	f.user.On("CreateUserNotify", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userDto.MSUserNotifyResponseDTO{}, nil)
	f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").Return(hfAuth, nil)
	f.consent.On("AcceptAll", mock.Anything, mock.Anything, "test-app", "corr-123").Return(userConsentDto.UserConsentAcceptAllResponseDTO{}, nil)
	f.auth.On("RegisterLogin", mock.Anything, mock.Anything, "test-app", "corr-123", mock.Anything).Return(hfLogin, nil)
	f.pub.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)
	f.consent.On("DeleteAllByUserId", mock.Anything, "user-123", "test-app", "corr-123").Return(nil)
	f.auth.On("ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123").Return(false, nil)
	return f
}

func newConfirmUC(f *confirmFixture, opts RegisterOptions) *registerConfirmUseCaseImpl {
	logger := zap.NewNop()
	return &registerConfirmUseCaseImpl{
		logger:              logger,
		userClient:          f.user,
		companyClient:       f.company,
		authClient:          f.auth,
		userConsentClient:   f.consent,
		communicationClient: f.comm,
		messagePublisher:    f.pub,
		sagaExecutor:        saga.NewInMemorySagaExecutor(logger),
		opts:                opts,
		lockTTL:             time.Minute,
		resultTTL:           time.Minute,
		waitTimeout:         2 * time.Second,
		pollInterval:        5 * time.Millisecond,
	}
}

// ---------------------------------------------------------------------------------------------------------------
// Confirm idempotente (flag ligada)
// ---------------------------------------------------------------------------------------------------------------

// Mesmo confirm repetido devolve o resultado guardado: a saga roda uma vez e o e-mail de boas-vindas sai uma vez.
func TestConfirmHardening_RepeatReturnsStoredResultWithoutSecondSaga(t *testing.T) {
	f := newConfirmFixture(nil)
	state, metrics := newFakeRegisterState(), newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state, Metrics: metrics})
	cmd := testConfirmCommand()

	first, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)
	second, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, "jwt-token-123", first.Token)
	assert.Equal(t, first, second)
	f.user.AssertNumberOfCalls(t, "ConfirmRegister", 1)
	f.auth.AssertNumberOfCalls(t, "CreateUser", 1)
	f.pub.AssertNumberOfCalls(t, "PublishMessage", 1) // só o boas-vindas do cadastro original
	assert.Equal(t, 1, metrics.step("ConfirmRegister", "replayed"))
	assert.Empty(t, state.locks, "trava deve sair ao fim da saga")
}

// Dois confirms simultâneos do mesmo token: um roda a saga, o outro espera e recebe o mesmo resultado.
func TestConfirmHardening_ConcurrentSameTokenOneWinsOtherGetsResult(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.user.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").
			Run(func(mock.Arguments) { time.Sleep(150 * time.Millisecond) }).Return(hfConfirm, nil)
	})
	state := newFakeRegisterState()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state, Metrics: newFakeRegistrationMetrics()})
	uc.waitTimeout = 5 * time.Second
	cmd := testConfirmCommand()

	type outcome struct {
		view appdto.RegisterConfirmViewDTO
		err  error
	}
	results := make(chan outcome, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := uc.Execute(context.Background(), cmd)
			results <- outcome{v, err}
		}()
	}
	wg.Wait()
	close(results)

	var got []outcome
	for r := range results {
		got = append(got, r)
	}
	require.Len(t, got, 2)
	for _, r := range got {
		require.NoError(t, r.err)
		assert.Equal(t, "jwt-token-123", r.view.Token)
	}
	f.user.AssertNumberOfCalls(t, "ConfirmRegister", 1)
	f.pub.AssertNumberOfCalls(t, "PublishMessage", 1)
}

// Outro conjunto de códigos não reaproveita o resultado: é outro confirm.
func TestConfirmHardening_DifferentCodesDoNotReplay(t *testing.T) {
	f := newConfirmFixture(nil)
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: newFakeRegisterState()})

	cmd := testConfirmCommand()
	_, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)

	other := cmd
	other.Token = "999999"
	_, err = uc.Execute(context.Background(), other)
	require.NoError(t, err)

	f.user.AssertNumberOfCalls(t, "ConfirmRegister", 2)
}

// Trava presa e sem resultado: o confirm concorrente desiste com 409 sem rodar a saga de novo.
func TestConfirmHardening_LockHeldWithoutResultReturnsInProgress(t *testing.T) {
	f := newConfirmFixture(nil)
	state := newFakeRegisterState()
	cmd := testConfirmCommand()
	state.locks[registerLockKeyPrefix+confirmFingerprint(confirmMaterial(cmd))] = true
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state})
	uc.waitTimeout = 30 * time.Millisecond

	_, err := uc.Execute(context.Background(), cmd)

	var appErr *pkg.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "REGISTRATION_IN_PROGRESS", appErr.Code)
	assert.Equal(t, 409, appErr.StatusCode)
	f.user.AssertNotCalled(t, "ConfirmRegister", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// Falha ao soltar o trava (saga falhou): a trava sai e não fica resultado; o retry roda a saga de novo.
func TestConfirmHardening_FailedSagaReleasesLockAndStoresNothing(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.consent.On("AcceptAll", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(userConsentDto.UserConsentAcceptAllResponseDTO{}, errors.New("consents fora"))
	})
	state := newFakeRegisterState()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state, Metrics: newFakeRegistrationMetrics()})
	cmd := testConfirmCommand()

	_, err := uc.Execute(context.Background(), cmd)
	require.Error(t, err)
	assert.False(t, state.lockHeld(registerLockKeyPrefix+confirmFingerprint(confirmMaterial(cmd))))
	assert.Empty(t, state.results)
}

// Redis fora do ar: segue sem trava (comportamento anterior) e conta o erro.
func TestConfirmHardening_StoreDownRunsWithoutLock(t *testing.T) {
	f := newConfirmFixture(nil)
	state := newFakeRegisterState()
	state.lockErr = errors.New("redis fora")
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state, Metrics: metrics})

	view, err := uc.Execute(context.Background(), testConfirmCommand())

	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", view.Token)
	f.user.AssertNumberOfCalls(t, "ConfirmRegister", 1)
	assert.Equal(t, 1, metrics.step("ConfirmLock", "store_error"))
}

// Flag desligada: nada de Redis, e confirm repetido roda a saga de novo (comportamento anterior).
func TestConfirmHardening_FlagOffNeverTouchesStateAndDoesNotReplay(t *testing.T) {
	f := newConfirmFixture(nil)
	state := newFakeRegisterState()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: false, State: state, Metrics: newFakeRegistrationMetrics()})
	cmd := testConfirmCommand()

	_, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)
	_, err = uc.Execute(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 0, state.tryLocks())
	f.user.AssertNumberOfCalls(t, "ConfirmRegister", 2)
}

// O resultado no Redis é cifrado com o material do cadastro: sem a sessão e os códigos, não abre; sem o token
// em texto plano.
func TestConfirmHardening_StoredResultIsEncryptedAndBoundToCodes(t *testing.T) {
	f := newConfirmFixture(nil)
	state := newFakeRegisterState()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, State: state})
	cmd := testConfirmCommand()

	_, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)

	material := confirmMaterial(cmd)
	resultKey := registerResultKeyPrefix + confirmFingerprint(material)
	sealed, ok := state.results[resultKey]
	require.True(t, ok, "resultado deve estar na chave com prefixo do produto")
	assert.False(t, bytes.Contains(sealed, []byte("jwt-token-123")), "token em texto plano no Redis")

	plain, err := openReplay(material, sealed)
	require.NoError(t, err)
	assert.Contains(t, string(plain), "jwt-token-123")

	otherCodes := cmd
	otherCodes.EmailToken = "000000"
	_, err = openReplay(confirmMaterial(otherCodes), sealed)
	assert.Error(t, err, "outros códigos não abrem o resultado")

	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 0xff
	_, err = openReplay(material, tampered)
	assert.Error(t, err, "dado adulterado não abre")

	assert.False(t, strings.Contains(resultKey, "123456"), "chave não pode conter código")
}

// Chave da trava e do resultado: mesmo material, mesma chave; outro material, outra chave.
func TestConfirmFingerprint_DependsOnSessionAndCodes(t *testing.T) {
	cmd := testConfirmCommand()
	base := confirmFingerprint(confirmMaterial(cmd))

	same := cmd
	same.Email = "outro@example.com" // e-mail não entra no material (a sessão já o identifica)
	assert.Equal(t, base, confirmFingerprint(confirmMaterial(same)))

	diffSession := cmd
	diffSession.RegistrationSessionId = "3f1c2a9e-0000-4000-8000-000000000002"
	assert.NotEqual(t, base, confirmFingerprint(confirmMaterial(diffSession)))

	diffCode := cmd
	diffCode.Token = "000001"
	assert.NotEqual(t, base, confirmFingerprint(confirmMaterial(diffCode)))
}

// ---------------------------------------------------------------------------------------------------------------
// Passo 5 com timeout ambíguo (flag ligada)
// ---------------------------------------------------------------------------------------------------------------

func timeoutErr() error { return fmt.Errorf("auth service: %w", context.DeadlineExceeded) }

// O ms-auth gravou a conta e a resposta se perdeu: a consulta acha a conta; o passo vira executado e a saga segue.
func TestCreateAuthUser_AmbiguousTimeoutButAccountExists_Continues(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, timeoutErr())
		f.auth.On("ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123").Return(true, nil)
	})
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, Metrics: metrics})

	view, err := uc.Execute(context.Background(), testConfirmCommand())

	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", view.Token)
	f.auth.AssertCalled(t, "ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123")
	assert.Equal(t, 1, metrics.step("CreateAuthUser", "ambiguous_recovered"))
}

// A conta não existe: o erro original segue e a saga compensa só o que foi executado (passo 3).
func TestCreateAuthUser_AmbiguousTimeoutNoAccount_FailsWithOriginalAndCompensates(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, timeoutErr())
	})
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, Metrics: metrics})

	_, err := uc.Execute(context.Background(), testConfirmCommand())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 1, metrics.comp("CreateUser", "delegated_to_reaper"))
}

// A consulta ao ms-auth também falha: fica com o erro original (não inventa sucesso).
func TestCreateAuthUser_AmbiguousTimeoutCheckFails_ReturnsOriginal(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, timeoutErr())
		f.auth.On("ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123").
			Return(false, errors.New("ms-auth 500"))
	})
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, Metrics: newFakeRegistrationMetrics()})

	_, err := uc.Execute(context.Background(), testConfirmCommand())
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// Flag desligada: timeout no passo 5 segue como antes (sem consulta ao ms-auth).
func TestCreateAuthUser_FlagOffTimeoutIsNotRecovered(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, timeoutErr())
	})
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: false})

	_, err := uc.Execute(context.Background(), testConfirmCommand())
	require.Error(t, err)
	f.auth.AssertNotCalled(t, "ExistsByIDExternal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// ---------------------------------------------------------------------------------------------------------------
// Compensação (sempre ligada, independe da flag)
// ---------------------------------------------------------------------------------------------------------------

// The saga no longer calls destructive routes of ms-auth or ms-user: HardDeleteUser and DeleteUser were removed
// from port.AuthClient/port.UserClient, so reintroducing them in the compensation does not compile. These tests
// pin the observable contract: delegated metric, no extra calls, original error preserved.

// Step 5 (CreateAuthUser) fails: step 3 cleanup is delegated to the ms-user reaper; nothing else is called.
func TestCompensation_Step5Failure_DelegatesToReaper(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, errors.New("ms-auth recusou"))
	})
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: false, Metrics: metrics})

	_, err := uc.Execute(context.Background(), testConfirmCommand())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ms-auth recusou")
	assert.Equal(t, 1, metrics.comp("CreateUser", "delegated_to_reaper"))
	assert.Equal(t, 0, metrics.comp("CreateUser", "ok"))
	assert.Equal(t, 0, metrics.comp("CreateAuthUser", "delegated_to_reaper"), "failed step is not compensated")
	f.consent.AssertNotCalled(t, "DeleteAllByUserId", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// Step 6 (AcceptConsents) fails: steps 5 and 3 are delegated; the original error is the only error returned.
func TestCompensation_Step6Failure_DelegatesToReaper(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.consent.On("AcceptAll", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(userConsentDto.UserConsentAcceptAllResponseDTO{}, errors.New("consents fora"))
	})
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, Metrics: metrics})

	_, err := uc.Execute(context.Background(), testConfirmCommand())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "consents fora")
	assert.NotContains(t, err.Error(), "compensa")
	assert.Equal(t, 1, metrics.comp("CreateAuthUser", "delegated_to_reaper"))
	assert.Equal(t, 1, metrics.comp("CreateUser", "delegated_to_reaper"))
	assert.Equal(t, 0, metrics.comp("CreateAuthUser", "error"))
	f.consent.AssertNotCalled(t, "DeleteAllByUserId", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// Step 7 (ExecuteLogin) fails after the client disconnects: the remaining compensation (consents, same company)
// still runs with its own non-cancelled context and deadline; ms-auth/ms-user stay delegated.
func TestCompensation_RemainingConsentCompensationStaysComplete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var consentCtxErr error
	var consentDeadline bool
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("RegisterLogin", mock.Anything, mock.Anything, "test-app", "corr-123", mock.Anything).
			Run(func(mock.Arguments) { cancel() }).
			Return(authDto.AuthLoginResponseDTO{}, errors.New("login fora"))
		f.consent.On("DeleteAllByUserId", mock.Anything, "user-123", "test-app", "corr-123").
			Run(func(a mock.Arguments) {
				c := a.Get(0).(context.Context)
				consentCtxErr = c.Err()
				_, consentDeadline = c.Deadline()
			}).Return(nil)
	})
	metrics := newFakeRegistrationMetrics()
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: false, Metrics: metrics})

	_, err := uc.Execute(ctx, testConfirmCommand())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "login fora")
	f.consent.AssertNumberOfCalls(t, "DeleteAllByUserId", 1)
	assert.NoError(t, consentCtxErr, "consent compensation got a cancelled context")
	assert.True(t, consentDeadline, "compensation without its own deadline")
	assert.Equal(t, 1, metrics.comp("AcceptConsents", "ok"))
	assert.Equal(t, 1, metrics.comp("CreateAuthUser", "delegated_to_reaper"))
	assert.Equal(t, 1, metrics.comp("CreateUser", "delegated_to_reaper"))
}

// Sem flag, sem Redis: o caminho de sucesso continua igual.
func TestConfirm_HappyPathUnchangedWithoutHardening(t *testing.T) {
	f := newConfirmFixture(nil)
	uc := newConfirmUC(f, RegisterOptions{})

	view, err := uc.Execute(context.Background(), testConfirmCommand())

	require.NoError(t, err)
	assert.Equal(t, "jwt-token-123", view.Token)
	assert.Equal(t, int64(3600), view.TokenExpiresIn)
}
