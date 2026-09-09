package user

import (
	"encoding/json"
	"strings"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestToMeProfile_DropsDocumentFields(t *testing.T) {
	raw := `{
		"email":"a@b.com",
		"display_handle":"handle",
		"personProfile":{"full_name":"Nome Completo","cpf":"123.456.789-00","rg":"1122233","mother_name":"Mae"}
	}`
	var user appdto.MSUserResponseDTO
	assert.NoError(t, json.Unmarshal([]byte(raw), &user))

	got := toMeProfile(user)
	encoded, err := json.Marshal(got)
	assert.NoError(t, err)
	payload := strings.ToLower(string(encoded))
	assert.NotContains(t, payload, "123.456.789-00")
	assert.NotContains(t, payload, "12345678900")
	assert.NotContains(t, payload, "1122233")
	assert.NotContains(t, payload, "mother")
	assert.Equal(t, "Nome Completo", got.PersonProfile.FullName)
	assert.True(t, got.PersonProfile.HasCpf)
	assert.Equal(t, "8900", got.PersonProfile.CpfLast4)
}

func TestToMeProfile_OmitsLast4WhenCpfMissing(t *testing.T) {
	user := appdto.MSUserResponseDTO{
		Email:         "a@b.com",
		PersonProfile: &appdto.PersonProfileDTO{FullName: "Nome Completo"},
	}
	got := toMeProfile(user)
	assert.Equal(t, "Nome Completo", got.PersonProfile.FullName)
	assert.False(t, got.PersonProfile.HasCpf)
	assert.Empty(t, got.PersonProfile.CpfLast4)
	encoded, err := json.Marshal(got)
	assert.NoError(t, err)
	assert.NotContains(t, strings.ToLower(string(encoded)), "123")
}
