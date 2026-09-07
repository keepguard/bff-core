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
	assert.NotContains(t, payload, "cpf")
	assert.NotContains(t, payload, "rg")
	assert.NotContains(t, payload, "mother")
	assert.Equal(t, "Nome Completo", got.PersonProfile.FullName)
}
