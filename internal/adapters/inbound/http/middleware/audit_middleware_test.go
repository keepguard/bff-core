package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keepguard/bff-core/internal/pkg"
	auditport "github.com/keepguard/bff-core/internal/domain/ports/audit"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type recPub struct {
	events []auditport.Event
}

func (r *recPub) Publish(_ context.Context, event auditport.Event) {
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

func TestShouldSkipProductGet(t *testing.T) {
	require.True(t, shouldSkipAudit(http.MethodGet, "/api/v1/core/collector/agents"))
	require.False(t, shouldSkipAudit(http.MethodGet, "/api/v1/core/audits"))
	require.False(t, shouldSkipAudit(http.MethodPost, "/api/v1/core/collector/agents"))
}
