package mapper

import (
	"encoding/json"
	"testing"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The public JSON of GET /users/me always carries countryCode: the ISO code or null.
func TestToMeProfileResponse_countryCode(t *testing.T) {
	pt := "PT"
	raw, err := json.Marshal(ToMeProfileResponse(appdto.MeProfileViewDTO{Email: "a@b.com", CountryCode: &pt}))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"countryCode":"PT"`)

	raw, err = json.Marshal(ToMeProfileResponse(appdto.MeProfileViewDTO{Email: "a@b.com"}))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"countryCode":null`)
}
