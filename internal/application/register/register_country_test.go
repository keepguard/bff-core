package register

import (
	"context"
	"encoding/json"
	"testing"

	userDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/user"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func sentInitRequest(t *testing.T, uc *registerInitUseCaseImpl) appdto.MSUserRegisterInitRequestDTO {
	t.Helper()
	calls := uc.userClient.(*MockUserClient).Calls
	require.Len(t, calls, 1)
	return calls[0].Arguments.Get(1).(appdto.MSUserRegisterInitRequestDTO)
}

// Achadinhos / InvestBot: no new fields → the body sent to the ms-user has no countryCode/preferredLocale keys.
func TestInit_withoutCountryKeepsTheSameBody(t *testing.T) {
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)
	uc := newInitUC(t, pub, RegisterOptions{}, zap.NewNop())
	_, err := uc.Execute(context.Background(), initCommand())
	require.NoError(t, err)

	raw, err := json.Marshal(sentInitRequest(t, uc))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "countryCode")
	assert.NotContains(t, string(raw), "preferredLocale")
}

func TestInit_passesCountryAndLocale(t *testing.T) {
	pub := &MockMessagePublisher{}
	pub.On("PublishMessage", mock.Anything, mock.Anything).Return(nil)
	uc := newInitUC(t, pub, RegisterOptions{}, zap.NewNop())
	cmd := initCommand()
	cmd.CountryCode, cmd.PreferredLocale = "PT", "pt-BR"
	_, err := uc.Execute(context.Background(), cmd)
	require.NoError(t, err)

	raw, err := json.Marshal(sentInitRequest(t, uc))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"countryCode":"PT","preferredLocale":"pt-BR"`)
}

func createUserRequest(t *testing.T, f *confirmFixture) appdto.MSUserCreateRequestDTO {
	t.Helper()
	for _, c := range f.user.Calls {
		if c.Method == "CreateUser" {
			return c.Arguments.Get(1).(appdto.MSUserCreateRequestDTO)
		}
	}
	t.Fatal("CreateUser not called")
	return appdto.MSUserCreateRequestDTO{}
}

// Confirm of a session without country/locale (Achadinhos, InvestBot, sessions from before): pt-BR as before and no
// countryCode key (the ms-user derives it).
func TestConfirm_withoutCountryUsesTodayDefaults(t *testing.T) {
	f := newConfirmFixture(nil)
	_, err := newConfirmUC(f, RegisterOptions{}).Execute(context.Background(), testConfirmCommand())
	require.NoError(t, err)
	req := createUserRequest(t, f)
	assert.Equal(t, "pt-BR", req.PreferredLocale)
	raw, _ := json.Marshal(req)
	assert.NotContains(t, string(raw), "countryCode")
}

func TestConfirm_createsUserWithCountryAndLocale(t *testing.T) {
	confirm := hfConfirm
	confirm.CountryCode, confirm.PreferredLocale = "PT", "en-US"
	f := newConfirmFixture(func(f *confirmFixture) {
		f.user.On("ConfirmRegister", mock.Anything, mock.Anything, "test-app", "corr-123").Return(confirm, nil)
	})
	_, err := newConfirmUC(f, RegisterOptions{}).Execute(context.Background(), testConfirmCommand())
	require.NoError(t, err)
	req := createUserRequest(t, f)
	assert.Equal(t, "PT", req.CountryCode)
	assert.Equal(t, "en-US", req.PreferredLocale)
}

// The ms-user confirm answer carries country_code / preferred_locale (snake_case, like the other session fields).
func TestConfirmResponse_decodesCountry(t *testing.T) {
	var r userDto.MSUserRegisterConfirmResponseDTO
	require.NoError(t, json.Unmarshal([]byte(`{"email":"a@b.c","country_code":"PT","preferred_locale":"pt-BR"}`), &r))
	assert.Equal(t, "PT", r.CountryCode)
	assert.Equal(t, "pt-BR", r.PreferredLocale)
}
