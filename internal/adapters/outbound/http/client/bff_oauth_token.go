package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	domainclient "github.com/keepguard/bff-core/internal/domain/ports/client"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/oauthsecret"
	"go.uber.org/zap"
)

const DefaultBffClientID = "bff-core"

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type tokenRequest struct {
	GrantType    string `json:"grantType"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type runtimeSecretResponse struct {
	ClientID        string `json:"clientId"`
	SecretEncrypted string `json:"secretEncrypted"`
	Status          string `json:"status"`
}

type tokenCacheEntry struct {
	token       string
	tokenExpiry time.Time
}

type secretCacheEntry struct {
	clientID string
	plain    string
}

type bffOAuthTokenHTTP struct {
	mu          sync.RWMutex
	cache       map[string]tokenCacheEntry
	secretCache map[string]secretCacheEntry
	renewBefore time.Duration
	clientID    string
	secretBase  string
	authBaseURL string
	httpClient  *http.Client
	logger      *zap.Logger
}

func NewBffOAuthTokenClient(cfg *config.Config, logger *zap.Logger) domainclient.ServiceTokenClient {
	renewBefore := time.Duration(cfg.OAuth.TokenRenewBeforeSec) * time.Second
	if renewBefore <= 0 {
		renewBefore = 10 * time.Minute
	}
	clientID := strings.TrimSpace(cfg.OAuth.ClientID)
	if clientID == "" {
		clientID = DefaultBffClientID
	}
	timeout := cfg.Services.Auth.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &bffOAuthTokenHTTP{
		cache:       make(map[string]tokenCacheEntry),
		secretCache: make(map[string]secretCacheEntry),
		renewBefore: renewBefore,
		clientID:    clientID,
		secretBase:  cfg.OAuth.SecretBase,
		authBaseURL: strings.TrimRight(cfg.Services.Auth.BaseURL, "/"),
		httpClient:  &http.Client{Timeout: timeout},
		logger:      logger,
	}
}

func (c *bffOAuthTokenHTTP) GetToken(ctx context.Context, companyID string) (string, error) {
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return "", fmt.Errorf("company_id é obrigatório para obter token OAuth do BFF")
	}

	c.mu.RLock()
	if entry, ok := c.cache[companyID]; ok && entry.token != "" && time.Until(entry.tokenExpiry) > c.renewBefore {
		token := entry.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.cache[companyID]; ok && entry.token != "" && time.Until(entry.tokenExpiry) > c.renewBefore {
		return entry.token, nil
	}

	plainSecret, clientID, err := c.plainSecretForCompany(ctx, companyID)
	if err != nil {
		return "", err
	}

	token, expiresIn, err := c.requestToken(ctx, companyID, clientID, plainSecret)
	if err != nil {
		return "", err
	}

	bearer := token
	if !strings.HasPrefix(strings.ToLower(bearer), "bearer ") {
		bearer = "Bearer " + token
	}
	c.cache[companyID] = tokenCacheEntry{
		token:       bearer,
		tokenExpiry: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	if c.logger != nil {
		c.logger.Debug("token OAuth do BFF renovado", zap.String("company_id", companyID))
	}
	return bearer, nil
}

func (c *bffOAuthTokenHTTP) plainSecretForCompany(ctx context.Context, companyID string) (string, string, error) {
	if entry, ok := c.secretCache[companyID]; ok && entry.plain != "" {
		return entry.plain, entry.clientID, nil
	}
	plain, clientID, err := c.fetchAndDecryptSecret(ctx, companyID)
	if err != nil {
		return "", "", err
	}
	c.secretCache[companyID] = secretCacheEntry{clientID: clientID, plain: plain}
	return plain, clientID, nil
}

func (c *bffOAuthTokenHTTP) fetchAndDecryptSecret(ctx context.Context, companyID string) (string, string, error) {
	query := url.Values{}
	query.Set("clientId", c.clientID)
	reqURL := c.authBaseURL + "/api/v1/auth/oauth/runtime/secret?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("falha ao criar request de ciphertext: %w", err)
	}
	req.Header.Set("X-Company-Id", companyID)
	req.Header.Set("X-Auth-Client-Secret-Base", c.secretBase)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("falha ao buscar ciphertext OAuth: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("falha ao ler ciphertext OAuth: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("ms-auth ciphertext status %d: %s", resp.StatusCode, string(respBody))
	}
	var material runtimeSecretResponse
	if err := json.Unmarshal(respBody, &material); err != nil {
		return "", "", fmt.Errorf("falha ao parsear ciphertext OAuth: %w", err)
	}
	if material.SecretEncrypted == "" {
		return "", "", fmt.Errorf("OAuth client bff-core sem secret cifrado para company %s; recrie o client", companyID)
	}
	plain, err := oauthsecret.DecryptAESGCM(c.secretBase, material.SecretEncrypted)
	if err != nil {
		return "", "", err
	}
	clientID := material.ClientID
	if clientID == "" {
		clientID = c.clientID
	}
	return plain, clientID, nil
}

func (c *bffOAuthTokenHTTP) requestToken(ctx context.Context, companyID, clientID, clientSecret string) (string, int64, error) {
	body, err := json.Marshal(tokenRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return "", 0, fmt.Errorf("falha ao serializar token request: %w", err)
	}

	tokenURL := c.authBaseURL + "/api/v1/auth/oauth/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("falha ao criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Company-Id", companyID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("falha na requisição de token: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("falha ao ler resposta: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("ms-auth retornou status %d: %s", resp.StatusCode, string(respBody))
	}
	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", 0, fmt.Errorf("falha ao parsear resposta de token: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", 0, fmt.Errorf("token vazio na resposta do ms-auth")
	}
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}
