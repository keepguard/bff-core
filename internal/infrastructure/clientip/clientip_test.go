package clientip

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromRequestXRealIPInternalCallerWins(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Chamador interno repassa o IP confiável via X-Real-IP
	req.Header.Set("X-Real-IP", "189.45.12.8")
	// XFF pode ter chegado com IPs forjados ou múltiplos
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 200.1.2.3")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "189.45.12.8", FromRequest(req))
}

func TestFromRequestXRealIPInternalPrivateWins(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Mesmo se for IP privado (ambiente interno/dev), X-Real-IP de chamador interno vence
	req.Header.Set("X-Real-IP", "10.42.0.50")
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 200.1.2.3")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "10.42.0.50", FromRequest(req))
}

func TestFromRequestXFFPicksLastPublicIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	// XFF com IP falso mandado pelo cliente na frente e IP real na borda atrás
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 200.1.2.3")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "200.1.2.3", FromRequest(req))
}

func TestFromRequestIgnoresCDNHeaders(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Headers de CDN que não existem na arquitetura sem CDN
	req.Header.Set("CF-Connecting-IP", "1.1.1.1")
	req.Header.Set("True-Client-IP", "2.2.2.2")
	req.RemoteAddr = "177.100.50.2:12345"

	assert.Equal(t, "177.100.50.2", FromRequest(req))
}

func TestFromRequestWithoutHeadersUsesRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "200.160.2.3:8080"

	assert.Equal(t, "200.160.2.3", FromRequest(req))
}

func TestFromRequestNilRequest(t *testing.T) {
	assert.Equal(t, "", FromRequest(nil))
}
