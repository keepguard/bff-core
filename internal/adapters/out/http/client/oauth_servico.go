package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// FonteOAuth pede ao ms-auth o token de serviço (client_credentials) deste serviço, em vez de
// assinar um ROLE_SYSTEM com o JWT_SECRET compartilhado. O client é próprio do serviço (revogável,
// com segredo rotacionável, auditado); quem invade o pod só leva o que esse client pode.
// Ambiente: KEEPGUARD_OAUTH_URL (padrão http://ms-auth:8081), KEEPGUARD_OAUTH_COMPANY_ID,
// KEEPGUARD_OAUTH_CLIENT_ID, KEEPGUARD_OAUTH_CLIENT_SECRET. Sem client e segredo, não há fonte
// (o serviço segue no modo antigo, em convivência). O token e o segredo nunca vão para log.
type FonteOAuth struct {
	url, companyID, clientID, secret string
	http                             *http.Client
	agora                            func() time.Time
	mu                               sync.Mutex
	atual                            string
	expira                           time.Time
}

const folgaOAuth = 60 * time.Second

// NovaFonteOAuth devolve nil se faltar company, client ou segredo.
func NovaFonteOAuth(url, companyID, clientID, segredo string) *FonteOAuth {
	url, companyID, clientID = strings.TrimRight(strings.TrimSpace(url), "/"), strings.TrimSpace(companyID), strings.TrimSpace(clientID)
	if url == "" || companyID == "" || clientID == "" || strings.TrimSpace(segredo) == "" {
		return nil
	}
	return &FonteOAuth{url: url, companyID: companyID, clientID: clientID, secret: segredo,
		http: &http.Client{Timeout: 5 * time.Second}, agora: time.Now}
}

// FonteOAuthDoAmbiente lê as variáveis KEEPGUARD_OAUTH_*; nil se o client não estiver configurado.
func FonteOAuthDoAmbiente() *FonteOAuth {
	url := os.Getenv("KEEPGUARD_OAUTH_URL")
	if strings.TrimSpace(url) == "" {
		url = "http://ms-auth:8081"
	}
	return NovaFonteOAuth(url, os.Getenv("KEEPGUARD_OAUTH_COMPANY_ID"), os.Getenv("KEEPGUARD_OAUTH_CLIENT_ID"),
		os.Getenv("KEEPGUARD_OAUTH_CLIENT_SECRET"))
}

// Token devolve o token vigente e renova quando faltar menos de 60 s.
func (f *FonteOAuth) Token() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	agora := f.agora()
	if f.atual != "" && f.expira.Sub(agora) > folgaOAuth {
		return f.atual, nil
	}
	corpo, _ := json.Marshal(map[string]string{"grantType": "client_credentials", "clientId": f.clientID, "clientSecret": f.secret})
	req, err := http.NewRequest(http.MethodPost, f.url+"/api/v1/auth/oauth/token", bytes.NewReader(corpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Company-Id", f.companyID)
	resp, err := f.http.Do(req)
	if err != nil {
		return "", errors.New("oauth: ms-auth indisponível")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth: ms-auth respondeu %d", resp.StatusCode)
	}
	var r struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int64  `json:"expiresIn"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&r); err != nil || r.AccessToken == "" {
		return "", errors.New("oauth: resposta ilegível do ms-auth")
	}
	f.atual = r.AccessToken
	f.expira = agora.Add(time.Duration(r.ExpiresIn) * time.Second)
	return f.atual, nil
}
