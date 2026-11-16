package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"github.com/keepguard/bff-core/internal/infrastructure/metrics"
	"github.com/keepguard/bff-core/internal/infrastructure/oauthsecret"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	testSecretBase = "test-base"
	companyA       = "aaaaaaaa-0000-0000-0000-00000000000a"
	companyB       = "bbbbbbbb-0000-0000-0000-00000000000b"
)

func fakeJWT(company, nonce string) string {
	enc := base64.RawURLEncoding
	payload, _ := json.Marshal(map[string]string{"company_id": company, "jti": nonce})
	return enc.EncodeToString([]byte(`{"alg":"HS256"}`)) + "." + enc.EncodeToString(payload) + ".sig"
}

// fakeCache is an in-test Redis stand-in (shared between client instances, like a real Redis).
type fakeCache struct {
	mu      sync.Mutex
	data    map[string]string
	ttl     map[string]time.Duration
	failAll bool
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: map[string]string{}, ttl: map[string]time.Duration{}}
}

var errFakeRedis = errors.New("redis down")

func (f *fakeCache) Get(_ context.Context, k string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failAll {
		return "", false, errFakeRedis
	}
	v, ok := f.data[k]
	return v, ok, nil
}
func (f *fakeCache) Set(_ context.Context, k, v string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failAll {
		return errFakeRedis
	}
	f.data[k], f.ttl[k] = v, ttl
	return nil
}
func (f *fakeCache) SetNX(_ context.Context, k string, ttl time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failAll {
		return false, errFakeRedis
	}
	if _, ok := f.data[k]; ok {
		return false, nil
	}
	f.data[k], f.ttl[k] = "1", ttl
	return true, nil
}
func (f *fakeCache) Del(_ context.Context, k string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, k)
	return nil
}

// fakeAuth simulates ms-auth: each company has its own client and secret; the token carries company_id.
type fakeAuth struct {
	srv          *httptest.Server
	tokenCalls   atomic.Int32
	secretCalls  atomic.Int32
	issueCompany func(header string) string // company placed in the token claim
	clientOf     map[string]string
	delay        time.Duration
}

func newFakeAuth(t *testing.T) *fakeAuth {
	t.Helper()
	fa := &fakeAuth{clientOf: map[string]string{companyA: "bff-core-a", companyB: "bff-core-b"}}
	fa.issueCompany = func(h string) string { return h }
	secrets := map[string]string{"bff-core-a": "secret-a", "bff-core-b": "secret-b"}
	fa.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		company := r.Header.Get("X-Company-Id")
		switch r.URL.Path {
		case "/api/v1/auth/oauth/runtime/secret":
			fa.secretCalls.Add(1)
			cid := r.URL.Query().Get("clientId")
			if fa.clientOf[company] != cid {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			enc, _ := oauthsecret.EncryptAESGCM(testSecretBase, secrets[cid])
			_ = json.NewEncoder(w).Encode(map[string]any{"clientId": cid, "secretEncrypted": enc, "status": "ACTIVE"})
		case "/api/v1/auth/oauth/token":
			n := fa.tokenCalls.Add(1)
			time.Sleep(fa.delay)
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if secrets[body["clientId"]] != body["clientSecret"] || fa.clientOf[company] != body["clientId"] {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"accessToken": fakeJWT(fa.issueCompany(company), fmt.Sprint(n)), "tokenType": "Bearer", "expiresIn": 3600,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(fa.srv.Close)
	return fa
}

func perCompanyCfg(baseURL string, ids map[string]string) *config.Config {
	return &config.Config{
		Services: config.ServicesConfig{Auth: config.ServiceConfig{BaseURL: baseURL}},
		OAuth: config.OAuthConfig{ClientID: "bff-core", SecretBase: testSecretBase, TokenRenewBeforeSec: 600,
			PerCompany: config.OAuthPerCompanyConfig{Enabled: true, ClientIDs: ids}},
	}
}

func newPC(t *testing.T, fa *fakeAuth, cache OAuthTokenCache) *perCompanyOAuthToken {
	t.Helper()
	c := NewPerCompanyOAuthTokenClient(perCompanyCfg(fa.srv.URL, fa.clientOf), cache, zap.NewNop()).(*perCompanyOAuthToken)
	c.waitStep = 5 * time.Millisecond
	return c
}

func TestPerCompany_EachCompanyGetsItsOwnToken(t *testing.T) {
	fa := newFakeAuth(t)
	c := newPC(t, fa, newFakeCache())

	tokA, err := c.GetToken(context.Background(), companyA)
	if err != nil {
		t.Fatal(err)
	}
	tokB, err := c.GetToken(context.Background(), companyB)
	if err != nil {
		t.Fatal(err)
	}
	if TokenCompanyID(tokA) != companyA || TokenCompanyID(tokB) != companyB || tokA == tokB {
		t.Fatal("token of one company used for another")
	}
}

// Critical: a token of company A sitting under company B's key is never returned for B.
func TestPerCompany_NeverServesAnotherCompanysToken(t *testing.T) {
	fa := newFakeAuth(t)
	cache := newFakeCache()
	c := newPC(t, fa, cache)
	tokA, err := c.GetToken(context.Background(), companyA)
	if err != nil {
		t.Fatal(err)
	}
	cache.data[oauthTokenKeyPrefix+companyB] = cache.data[oauthTokenKeyPrefix+companyA] // poisoned entry
	tokB, err := c.GetToken(context.Background(), companyB)
	if err != nil {
		t.Fatal(err)
	}
	if tokB == tokA || TokenCompanyID(tokB) != companyB {
		t.Fatal("company B received company A's token")
	}
}

func TestPerCompany_NoClientForCompany_ClearErrorNoFallback(t *testing.T) {
	fa := newFakeAuth(t)
	c := NewPerCompanyOAuthTokenClient(perCompanyCfg(fa.srv.URL, map[string]string{companyA: "bff-core-a"}), newFakeCache(), zap.NewNop())
	before := metrics.OAuthTokenCount(metrics.OAuthResultNoClient)

	_, err := c.GetToken(context.Background(), companyB)
	if !errors.Is(err, ErrNoCompanyClient) {
		t.Fatalf("want ErrNoCompanyClient, got %v", err)
	}
	if fa.secretCalls.Load() != 0 || fa.tokenCalls.Load() != 0 {
		t.Fatal("must not call ms-auth (and never with another company's client)")
	}
	if metrics.OAuthTokenCount(metrics.OAuthResultNoClient) != before+1 {
		t.Fatal("no_client metric not counted")
	}
}

func TestPerCompany_TokenIssuedForAnotherCompany_Rejected(t *testing.T) {
	fa := newFakeAuth(t)
	fa.issueCompany = func(string) string { return companyA }
	cache := newFakeCache()
	c := newPC(t, fa, cache)

	_, err := c.GetToken(context.Background(), companyB)
	if !errors.Is(err, ErrTokenCompanyMismatch) {
		t.Fatalf("want mismatch, got %v", err)
	}
	if _, ok := cache.data[oauthTokenKeyPrefix+companyB]; ok {
		t.Fatal("mismatched token must not be cached")
	}
}

// Cache only in Redis: a second instance (another pod) reuses the Redis entry; with Redis emptied it fetches again,
// which proves there is no copy in process memory.
func TestPerCompany_NoProcessMemoryCache(t *testing.T) {
	fa := newFakeAuth(t)
	cache := newFakeCache()
	c1 := newPC(t, fa, cache)
	if _, err := c1.GetToken(context.Background(), companyA); err != nil {
		t.Fatal(err)
	}
	c2 := newPC(t, fa, cache)
	if _, err := c2.GetToken(context.Background(), companyA); err != nil {
		t.Fatal(err)
	}
	if fa.tokenCalls.Load() != 1 {
		t.Fatalf("want 1 token call (shared Redis), got %d", fa.tokenCalls.Load())
	}
	delete(cache.data, oauthTokenKeyPrefix+companyA)
	if _, err := c1.GetToken(context.Background(), companyA); err != nil {
		t.Fatal(err)
	}
	if fa.tokenCalls.Load() != 2 {
		t.Fatalf("process kept a token in memory: token calls = %d", fa.tokenCalls.Load())
	}
}

func TestPerCompany_CachedValueEncryptedAndTTLBelowExpiry(t *testing.T) {
	fa := newFakeAuth(t)
	cache := newFakeCache()
	c := newPC(t, fa, cache)
	tok, err := c.GetToken(context.Background(), companyA)
	if err != nil {
		t.Fatal(err)
	}
	key := oauthTokenKeyPrefix + companyA
	raw := strings.TrimPrefix(tok, "Bearer ")
	if strings.Contains(cache.data[key], raw) {
		t.Fatal("token stored in clear in Redis")
	}
	if ttl := cache.ttl[key]; ttl <= 0 || ttl >= time.Hour {
		t.Fatalf("ttl %v must be > 0 and < token expiry", ttl)
	}
	if !strings.HasPrefix(key, "kg:bff-core:oauth:token:") {
		t.Fatal("key prefix")
	}
}

func TestPerCompany_RedisDown_FetchesDirectly(t *testing.T) {
	fa := newFakeAuth(t)
	cache := newFakeCache()
	cache.failAll = true
	c := newPC(t, fa, cache)
	before := metrics.OAuthTokenCount(metrics.OAuthResultCacheBypass)
	tok, err := c.GetToken(context.Background(), companyA)
	if err != nil || TokenCompanyID(tok) != companyA {
		t.Fatalf("want direct token, got err=%v", err)
	}
	if metrics.OAuthTokenCount(metrics.OAuthResultCacheBypass) != before+1 {
		t.Fatal("cache_bypass not counted")
	}
}

func TestPerCompany_SingleFlightPerCompany(t *testing.T) {
	fa := newFakeAuth(t)
	fa.delay = 30 * time.Millisecond
	cache := newFakeCache()
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := newPC(t, fa, cache)
			tok, err := c.GetToken(context.Background(), companyA)
			if err == nil && TokenCompanyID(tok) != companyA {
				err = errors.New("wrong company")
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if fa.tokenCalls.Load() != 1 {
		t.Fatalf("burst made %d token calls, want 1", fa.tokenCalls.Load())
	}
	if _, locked := cache.data[oauthLockKeyPrefix+companyA]; locked {
		t.Fatal("lock not released")
	}
}

func TestCacheTTL(t *testing.T) {
	if cacheTTL(time.Hour, 10*time.Minute) != 50*time.Minute || cacheTTL(time.Minute, 10*time.Minute) != 30*time.Second || cacheTTL(0, 0) != 0 {
		t.Fatal("cacheTTL")
	}
}

// Integration with a real Redis (REDIS_ADDR, e.g. 127.0.0.1:6399). Skipped without it.
func TestPerCompany_RealRedis(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}
	rc := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	defer rc.Close()
	ctx := context.Background()
	if err := rc.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unreachable: %v", err)
	}
	_ = rc.Del(ctx, oauthTokenKeyPrefix+companyA, oauthTokenKeyPrefix+companyB, oauthLockKeyPrefix+companyA).Err()
	fa := newFakeAuth(t)
	cache := NewRedisOAuthTokenCache(rc)
	c1 := newPC(t, fa, cache)
	c2 := newPC(t, fa, cache)
	tokA, err := c1.GetToken(ctx, companyA)
	if err != nil {
		t.Fatal(err)
	}
	tokA2, err := c2.GetToken(ctx, companyA)
	if err != nil || tokA2 != tokA || fa.tokenCalls.Load() != 1 {
		t.Fatalf("redis cache not shared: err=%v calls=%d", err, fa.tokenCalls.Load())
	}
	ttl := rc.TTL(ctx, oauthTokenKeyPrefix+companyA).Val()
	if ttl <= 0 || ttl >= time.Hour {
		t.Fatalf("ttl %v", ttl)
	}
	tokB, err := c1.GetToken(ctx, companyB)
	if err != nil || TokenCompanyID(tokB) != companyB {
		t.Fatal("company B token")
	}
	_ = rc.Del(ctx, oauthTokenKeyPrefix+companyA, oauthTokenKeyPrefix+companyB).Err()
}
