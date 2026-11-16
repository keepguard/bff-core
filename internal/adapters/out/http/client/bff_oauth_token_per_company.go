package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainclient "github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/infrastructure/oauthsecret"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis keys of the per-company service token (W5). Nothing is kept in process memory.
const (
	oauthTokenKeyPrefix = "kg:bff-core:oauth:token:"
	oauthLockKeyPrefix  = "kg:bff-core:oauth:lock:"
	oauthLockTTL        = 5 * time.Second
	oauthWaitStep       = 100 * time.Millisecond
	oauthWaitSteps      = 20
)

var (
	// ErrNoCompanyClient: the company has no OAuth client of its own; never falls back to another company's client.
	ErrNoCompanyClient = errors.New("no oauth client configured for company")
	// ErrTokenCompanyMismatch: the token (or the secret material) does not belong to the requested company.
	ErrTokenCompanyMismatch = errors.New("oauth token company mismatch")
)

// OAuthTokenCache is the Redis surface used by the per-company token client.
type OAuthTokenCache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Del(ctx context.Context, key string) error
}

type redisOAuthTokenCache struct{ client *redis.Client }

// NewRedisOAuthTokenCache wraps a go-redis client. A nil client makes every call fail (callers bypass the cache).
func NewRedisOAuthTokenCache(client *redis.Client) OAuthTokenCache {
	return &redisOAuthTokenCache{client: client}
}

var errOAuthCacheUnavailable = errors.New("redis unavailable for oauth token cache")

func (r *redisOAuthTokenCache) Get(ctx context.Context, key string) (string, bool, error) {
	if r.client == nil {
		return "", false, errOAuthCacheUnavailable
	}
	v, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (r *redisOAuthTokenCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if r.client == nil {
		return errOAuthCacheUnavailable
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisOAuthTokenCache) SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r.client == nil {
		return false, errOAuthCacheUnavailable
	}
	return r.client.SetNX(ctx, key, "1", ttl).Result()
}

func (r *redisOAuthTokenCache) Del(ctx context.Context, key string) error {
	if r.client == nil {
		return errOAuthCacheUnavailable
	}
	return r.client.Del(ctx, key).Err()
}

// perCompanyOAuthToken obtains the bff-core service token with the company's own client. The token is cached
// in Redis encrypted (AES-GCM with the secret base), with a TTL shorter than its expiry. Tokens are never logged.
type perCompanyOAuthToken struct {
	fetcher     *bffOAuthTokenHTTP // only its HTTP helpers are used; its in-memory maps are not touched
	clientIDs   map[string]string
	cache       OAuthTokenCache
	renewBefore time.Duration
	secretBase  string
	logger      *zap.Logger
	waitStep    time.Duration
}

// NewPerCompanyOAuthTokenClient builds the W5 token client (oauth.per_company.enabled).
func NewPerCompanyOAuthTokenClient(cfg *config.Config, cache OAuthTokenCache, logger *zap.Logger) domainclient.ServiceTokenClient {
	base := NewBffOAuthTokenClient(cfg, logger).(*bffOAuthTokenHTTP)
	ids := make(map[string]string, len(cfg.OAuth.PerCompany.ClientIDs))
	for company, client := range cfg.OAuth.PerCompany.ClientIDs {
		company, client = strings.ToLower(strings.TrimSpace(company)), strings.TrimSpace(client)
		if company != "" && client != "" {
			ids[company] = client
		}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &perCompanyOAuthToken{
		fetcher:     base,
		clientIDs:   ids,
		cache:       cache,
		renewBefore: base.renewBefore,
		secretBase:  cfg.OAuth.SecretBase,
		logger:      logger,
		waitStep:    oauthWaitStep,
	}
}

func (c *perCompanyOAuthToken) GetToken(ctx context.Context, companyID string) (string, error) {
	companyID = strings.ToLower(strings.TrimSpace(companyID))
	if companyID == "" {
		return "", fmt.Errorf("company_id is required for the bff-core service token")
	}
	clientID, ok := c.clientIDs[companyID]
	if !ok {
		metrics.RecordOAuthToken(metrics.OAuthResultNoClient)
		c.logger.Warn("no per-company oauth client", zap.String("company_id", companyID))
		return "", fmt.Errorf("%w: %s", ErrNoCompanyClient, companyID)
	}

	key := oauthTokenKeyPrefix + companyID
	if tok, hit, err := c.readCache(ctx, key, companyID); err != nil {
		metrics.RecordOAuthToken(metrics.OAuthResultCacheBypass)
		c.logger.Warn("oauth token cache unavailable; fetching without cache", zap.String("company_id", companyID), zap.Error(err))
		return c.fetch(ctx, companyID, clientID, "")
	} else if hit {
		metrics.RecordOAuthToken(metrics.OAuthResultCacheHit)
		return tok, nil
	}

	// Single-flight per company: one fetch at a time; the others wait for the cache to fill.
	lockKey := oauthLockKeyPrefix + companyID
	got, err := c.cache.SetNX(ctx, lockKey, oauthLockTTL)
	if err != nil {
		metrics.RecordOAuthToken(metrics.OAuthResultCacheBypass)
		return c.fetch(ctx, companyID, clientID, "")
	}
	if got {
		defer func() { _ = c.cache.Del(context.WithoutCancel(ctx), lockKey) }()
		return c.fetch(ctx, companyID, clientID, key)
	}
	for i := 0; i < oauthWaitSteps; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(c.waitStep):
		}
		if tok, hit, err := c.readCache(ctx, key, companyID); err == nil && hit {
			metrics.RecordOAuthToken(metrics.OAuthResultCacheHit)
			return tok, nil
		}
	}
	return c.fetch(ctx, companyID, clientID, key)
}

// readCache returns the cached bearer when present, decryptable and bound to the same company.
func (c *perCompanyOAuthToken) readCache(ctx context.Context, key, companyID string) (string, bool, error) {
	raw, ok, err := c.cache.Get(ctx, key)
	if err != nil || !ok {
		return "", false, err
	}
	plain, derr := oauthsecret.DecryptAESGCM(c.secretBase, raw)
	if derr != nil {
		return "", false, nil
	}
	if TokenCompanyID(plain) != companyID {
		metrics.RecordOAuthToken(metrics.OAuthResultCompanyMismatch)
		return "", false, nil
	}
	return plain, true, nil
}

// fetch gets the secret of the company's client and a fresh token; cacheKey "" = do not cache.
func (c *perCompanyOAuthToken) fetch(ctx context.Context, companyID, clientID, cacheKey string) (string, error) {
	plain, gotClient, err := c.fetcher.fetchAndDecryptSecret(ctx, companyID, clientID)
	if err != nil {
		metrics.RecordOAuthToken(metrics.OAuthResultError)
		return "", err
	}
	if gotClient != clientID {
		metrics.RecordOAuthToken(metrics.OAuthResultCompanyMismatch)
		return "", fmt.Errorf("%w: secret material of another client", ErrTokenCompanyMismatch)
	}
	token, expiresIn, err := c.fetcher.requestToken(ctx, companyID, clientID, plain)
	if err != nil {
		metrics.RecordOAuthToken(metrics.OAuthResultError)
		return "", err
	}
	bearer := token
	if !strings.HasPrefix(strings.ToLower(bearer), "bearer ") {
		bearer = "Bearer " + token
	}
	if TokenCompanyID(bearer) != companyID {
		metrics.RecordOAuthToken(metrics.OAuthResultCompanyMismatch)
		c.logger.Warn("service token issued for another company; discarded", zap.String("company_id", companyID))
		return "", fmt.Errorf("%w: %s", ErrTokenCompanyMismatch, companyID)
	}
	metrics.RecordOAuthToken(metrics.OAuthResultFetched)
	if cacheKey != "" {
		if ttl := cacheTTL(time.Duration(expiresIn)*time.Second, c.renewBefore); ttl > 0 {
			if enc, eerr := oauthsecret.EncryptAESGCM(c.secretBase, bearer); eerr == nil {
				_ = c.cache.Set(ctx, cacheKey, enc, ttl)
			}
		}
	}
	return bearer, nil
}

// cacheTTL is always shorter than the token expiry: expiry - renewBefore, or half the expiry when that is too short.
func cacheTTL(expiry, renewBefore time.Duration) time.Duration {
	if expiry <= 0 {
		return 0
	}
	if ttl := expiry - renewBefore; ttl > 0 {
		return ttl
	}
	return expiry / 2
}

// TokenCompanyID reads the company_id claim of a JWT (with or without "Bearer "), without verifying the
// signature (ms-auth verifies it). Returns "" when absent or unreadable.
func TokenCompanyID(token string) string {
	token = strings.TrimSpace(token)
	if len(token) > 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return ""
	}
	var claims struct {
		CompanyID string `json:"company_id"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(claims.CompanyID))
}
