package clientip

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromRequestDiscardsSpoofedPublicIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Public-IP", "189.45.12.8")
	req.Header.Set("X-Client-IP", "189.45.12.9")
	req.Header.Set("X-Forwarded-For", "200.100.50.1")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "200.100.50.1", FromRequest(req))
}

func TestFromRequestPrefersPublicClientIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.42.0.1, 189.45.12.8")
	req.Header.Set("X-Real-IP", "10.42.0.1")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "189.45.12.8", FromRequest(req))
}

func TestFromRequestUsesCFConnectingIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("CF-Connecting-IP", "200.160.2.3")
	req.Header.Set("X-Forwarded-For", "10.42.0.1")
	req.RemoteAddr = "10.42.0.15:443"

	assert.Equal(t, "200.160.2.3", FromRequest(req))
}

func TestFromRequestFallsBackToPrivate(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.42.0.1")
	req.RemoteAddr = "127.0.0.1:8080"

	assert.Equal(t, "10.42.0.1", FromRequest(req))
}
