package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestFonteOAuth_PedeCacheiaERenova(t *testing.T) {
	var chamadas atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas.Add(1)
		var in map[string]string
		_ = json.NewDecoder(r.Body).Decode(&in)
		if r.URL.Path != "/api/v1/auth/oauth/token" || r.Header.Get("X-Company-Id") != "co-1" ||
			in["grantType"] != "client_credentials" || in["clientId"] != "srv" || in["clientSecret"] != "s3" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"accessToken": "tok" + string(rune('0'+chamadas.Load())), "tokenType": "Bearer", "expiresIn": 300})
	}))
	defer srv.Close()
	agora := time.Now()
	f := NovaFonteOAuth(srv.URL, "co-1", "srv", "s3")
	f.agora = func() time.Time { return agora }
	a, err := f.Token()
	if err != nil || a != "tok1" {
		t.Fatalf("1º token = %q %v", a, err)
	}
	if b, _ := f.Token(); b != "tok1" || chamadas.Load() != 1 {
		t.Fatalf("o 2º uso vem do cache: %q chamadas=%d", b, chamadas.Load())
	}
	agora = agora.Add(250 * time.Second) // faltam 50 s: renova
	if c, _ := f.Token(); c != "tok2" || chamadas.Load() != 2 {
		t.Fatalf("renovação: %q chamadas=%d", c, chamadas.Load())
	}
}

func TestFonteOAuth_ErroNaoVazaSegredoESemConfigEhNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer srv.Close()
	f := NovaFonteOAuth(srv.URL, "co-1", "srv", "segredo-super-secreto")
	_, err := f.Token()
	if err == nil || contem(err.Error(), "segredo-super-secreto") {
		t.Fatalf("erro = %v", err)
	}
	if NovaFonteOAuth("", "", "", "") != nil || NovaFonteOAuth(srv.URL, "c", "srv", "") != nil {
		t.Fatal("sem client/segredo/company não há fonte")
	}
}

func contem(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
