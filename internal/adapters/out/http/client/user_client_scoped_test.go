package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
	"go.uber.org/zap"
)

type tokensByCompany map[string]string

func (t tokensByCompany) GetToken(_ context.Context, company string) (string, error) {
	if tok, ok := t[company]; ok {
		return tok, nil
	}
	return "", ErrNoCompanyClient
}

type upstreamRec struct {
	calls   atomic.Int32
	auth    atomic.Value
	company atomic.Value
}

func newScopedUpstream(t *testing.T, body string) (*httptest.Server, *upstreamRec) {
	rec := &upstreamRec{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.calls.Add(1)
		rec.auth.Store(r.Header.Get("Authorization"))
		rec.company.Store(r.Header.Get("X-Company-Id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

func scopedClient(baseURL string, tokens port.ServiceTokenClient) *userClient {
	cfg := &config.Config{}
	cfg.Services.User.BaseURL = baseURL
	cfg.Services.Auth.BaseURL = baseURL
	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	WithCompanyScopedCalls(c, tokens)
	return c
}

func TestScopedPatch_UsesTargetCompanyToken(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{"id":"u1","companyId":"`+companyA+`"}`)
	tokA := "Bearer " + fakeJWT(companyA, "1")
	c := scopedClient(srv.URL, tokensByCompany{companyA: tokA, companyB: "Bearer " + fakeJWT(companyB, "2")})
	ctx := port.WithTargetCompanyID(port.WithCompanyID(context.Background(), companyA), companyA)

	if _, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "person-token", "tenant", "corr"); err != nil {
		t.Fatal(err)
	}
	if rec.auth.Load() != tokA || rec.company.Load() != companyA {
		t.Fatal("must send company A token with X-Company-Id A")
	}
}

func TestScopedPatch_TargetOfAnotherCompany_RefusedBeforeCall(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{"id":"u1"}`)
	c := scopedClient(srv.URL, tokensByCompany{companyA: "Bearer " + fakeJWT(companyA, "1"), companyB: "Bearer " + fakeJWT(companyB, "2")})
	for name, ctx := range map[string]context.Context{
		"other company": port.WithTargetCompanyID(port.WithCompanyID(context.Background(), companyA), companyB),
		"unknown":       port.WithCompanyID(context.Background(), companyA),
	} {
		_, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "person-token", "tenant", "corr")
		if !errors.Is(err, ErrCrossCompanyCall) {
			t.Fatalf("%s: want ErrCrossCompanyCall, got %v", name, err)
		}
	}
	if rec.calls.Load() != 0 {
		t.Fatal("ms-user must not be called")
	}
}

// The token provider returns a token bound to another company: refused (X-Company-Id must equal token company).
func TestScoped_TokenOfAnotherCompany_Refused(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{"id":"u1"}`)
	c := scopedClient(srv.URL, tokensByCompany{companyA: "Bearer " + fakeJWT(companyB, "x")})
	ctx := port.WithTargetCompanyID(port.WithCompanyID(context.Background(), companyA), companyA)
	if _, err := c.PatchPersonDocument(ctx, "u1", "52998224725", "", "t", "c"); !errors.Is(err, ErrCrossCompanyCall) {
		t.Fatalf("patch: %v", err)
	}
	if _, err := c.GetByEmail(ctx, "x@example.com", "t", companyA, "c"); !errors.Is(err, ErrCrossCompanyCall) {
		t.Fatalf("get: %v", err)
	}
	if rec.calls.Load() != 0 {
		t.Fatal("no call expected")
	}
}

func TestScopedGetByEmail_OtherCompanyRefusedBeforeCall(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{"id":"1","codeUser":""}`)
	c := scopedClient(srv.URL, tokensByCompany{companyA: "Bearer " + fakeJWT(companyA, "1"), companyB: "Bearer " + fakeJWT(companyB, "2")})
	ctx := port.WithCompanyID(context.Background(), companyA)
	if _, err := c.GetByEmail(ctx, "x@example.com", "t", companyB, "c"); !errors.Is(err, ErrCrossCompanyCall) {
		t.Fatalf("want refusal, got %v", err)
	}
	if rec.calls.Load() != 0 {
		t.Fatal("no call expected")
	}
	if _, err := c.GetByEmail(ctx, "x@example.com", "t", companyA, "c"); err != nil {
		t.Fatal(err)
	}
	if rec.company.Load() != companyA || TokenCompanyID(rec.auth.Load().(string)) != companyA {
		t.Fatal("must use company A token and header")
	}
}

func TestScopedGetByEmail_NoClientForCompany_Error(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{}`)
	c := scopedClient(srv.URL, tokensByCompany{})
	ctx := port.WithCompanyID(context.Background(), companyA)
	if _, err := c.GetByEmail(ctx, "x@example.com", "t", companyA, "c"); !errors.Is(err, ErrNoCompanyClient) {
		t.Fatalf("want ErrNoCompanyClient, got %v", err)
	}
	if rec.calls.Load() != 0 {
		t.Fatal("no call expected")
	}
}

// Flag off: previous behavior (no service token on GetByEmail, cross-company not checked).
func TestUnscoped_PreviousBehavior(t *testing.T) {
	srv, rec := newScopedUpstream(t, `{"id":"1","codeUser":""}`)
	cfg := &config.Config{}
	cfg.Services.User.BaseURL, cfg.Services.Auth.BaseURL = srv.URL, srv.URL
	c := &userClient{httpClient: resty.New(), config: cfg, logger: zap.NewNop()}
	ctx := port.WithCompanyID(context.Background(), companyA)
	if _, err := c.GetByEmail(ctx, "x@example.com", "t", companyB, "c"); err != nil {
		t.Fatal(err)
	}
	if rec.auth.Load() != "" || rec.company.Load() != companyB {
		t.Fatal("flag off must keep the old call")
	}
}
