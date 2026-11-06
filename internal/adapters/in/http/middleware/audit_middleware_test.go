package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	outport "github.com/keepguard/bff-core/internal/application/port/out"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type recPub struct {
	events []outport.Event
}

func (r *recPub) Publish(_ context.Context, event outport.Event) {
	r.events = append(r.events, event)
}

func (r *recPub) Close() error { return nil }

func TestMapAuditActionRegister(t *testing.T) {
	require.Equal(t, "REGISTER_INIT", mapAuditAction(http.MethodPost, "/api/v1/register/init"))
	require.Equal(t, "REGISTER_CONFIRM", mapAuditAction(http.MethodPost, "/api/v1/register/confirm"))
	require.Equal(t, "ACCEPT_CONSENTS", mapAuditAction(http.MethodPost, "/api/v1/user-consents/accept-batch"))
}

func TestAuditMiddlewareSkipsRegisterSuccess(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", nil)
	req.Header.Set("X-Correlation-ID", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/register/init")

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	require.NoError(t, err)
	require.Empty(t, rec.events)
}

func TestAuditMiddlewareSkipsRegisterFailure(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register/init", nil)
	req.Header.Set("X-Correlation-ID", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/register/init")

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusBadRequest)
	})(c)
	require.NoError(t, err)
	require.Empty(t, rec.events)
}

func TestAuditMiddlewareSkipsCollectorSuccess(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/collector/agents", nil)
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/core/collector/agents")
	c.Set("claims", &pkg.JWTClaims{CodeUser: "u-1", TenantId: "t-1"})

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})(c)
	require.NoError(t, err)
	require.Empty(t, rec.events)
}

func TestAuditMiddlewareSkipsOAuthWriteSuccess(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/oauth/clients", nil)
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/core/oauth/clients")
	c.Set("claims", &pkg.JWTClaims{CodeUser: "u-1", TenantId: "t-1"})

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})(c)
	require.NoError(t, err)
	require.Empty(t, rec.events)
}

func TestAuditMiddlewareDeniedWithoutJWTOnProtectedWrite(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/core/oauth/clients", nil)
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/core/oauth/clients")

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusUnauthorized)
	})(c)
	require.NoError(t, err)
	require.Len(t, rec.events, 1)
	require.Equal(t, "DENIED", rec.events[0].Outcome)
	require.Equal(t, "ANONYMOUS", rec.events[0].Actor.Type)
}

func TestAuditMiddlewareAuditReadPublishes(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/core/audits/evt-1", nil)
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/core/audits/:eventId")
	c.SetParamNames("eventId")
	c.SetParamValues("evt-1")
	c.Set("claims", &pkg.JWTClaims{CodeUser: "u-1", TenantId: "t-1"})

	mw := AuditMiddleware(rec, "bff-core")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	require.NoError(t, err)
	require.Len(t, rec.events, 1)
	require.Equal(t, "AUDIT_READ", rec.events[0].Action)
	require.Equal(t, "u-1", rec.events[0].Actor.CodeUser)
	require.Equal(t, "t-1", rec.events[0].TenantID)
	require.Equal(t, "evt-1", rec.events[0].Resource.ID)
}

// §13 e §11: evento de auditoria com IP truncado e sem aparelho, mesmo com o aparelho no
// token e no cabeçalho. O evento serializado também não leva o IP completo.
func TestAuditMiddlewareTruncatesIPAndDropsDevice(t *testing.T) {
	cases := []struct {
		name   string
		realIP string
		want   string
	}{
		{"IPv4", "203.0.113.77", "203.0.113.0"},
		{"IPv6", "2001:db8:1:2:3:4:5:6", "2001:db8:1::"},
		{"cabecalho que nao e IP", "nao-e-ip", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &recPub{}
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/credits", nil)
			req.Header.Set("X-Real-IP", tc.realIP)
			req.Header.Set("X-Device-Id", "dev-cabecalho")
			rr := httptest.NewRecorder()
			c := e.NewContext(req, rr)
			c.SetPath("/api/v1/core/billing/credits")
			c.Set("claims", &pkg.JWTClaims{CodeUser: "u-1", TenantId: "t-1", DeviceID: "dev-token"})

			err := AuditMiddleware(rec, "bff-core")(func(c echo.Context) error {
				return c.NoContent(http.StatusForbidden)
			})(c)
			require.NoError(t, err)
			require.Len(t, rec.events, 1)
			require.Equal(t, tc.want, rec.events[0].Actor.ClientIP)
			require.Empty(t, rec.events[0].Actor.DeviceID)
			raw, err := json.Marshal(rec.events[0])
			require.NoError(t, err)
			require.NotContains(t, string(raw), "dev-")
			require.NotContains(t, string(raw), "deviceId")
			require.NotContains(t, string(raw), tc.realIP)
		})
	}
}

func TestShouldSkipProductGet(t *testing.T) {
	require.True(t, shouldSkipAudit(http.MethodGet, "/api/v1/core/collector/agents"))
	require.True(t, shouldSkipAudit(http.MethodGet, "/api/v1/core/audits"), "listagem de auditoria deve ser skip para evitar auto-auditoria em looping")
	require.False(t, shouldSkipAudit(http.MethodGet, "/api/v1/core/audits/evt-1"), "leitura de evento específico deve ser auditada")
	require.False(t, shouldSkipAudit(http.MethodPost, "/api/v1/core/collector/agents"))
}
