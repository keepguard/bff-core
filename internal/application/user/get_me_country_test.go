package user

import (
	"encoding/json"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GET /users/me: countryCode from the ms-user (string), or nil when the ms-user has none.
func TestMe_countryCode(t *testing.T) {
	var user appdto.MSUserResponseDTO
	require.NoError(t, json.Unmarshal([]byte(`{"email":"a@b.com","countryCode":"PT"}`), &user))
	got := toMeProfile(user)
	require.NotNil(t, got.CountryCode)
	assert.Equal(t, "PT", *got.CountryCode)

	var legacy appdto.MSUserResponseDTO
	require.NoError(t, json.Unmarshal([]byte(`{"email":"a@b.com","countryCode":null}`), &legacy))
	assert.Nil(t, toMeProfile(legacy).CountryCode)
}
