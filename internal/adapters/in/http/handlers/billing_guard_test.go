package handlers

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	appbilling "github.com/keepguard/bff-core/internal/application/billing"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// noCallBilling é o cliente do billing que não pode ser chamado: qualquer método conta e entra em pânico.
type noCallBilling struct {
	port.BillingClient
	calls *int32
}

// billingGuardHandlers são todos os handlers de billing que passam pelo guard (os webhooks não passam).
func billingGuardHandlers(h *BillingHandlers) map[string]func(echo.Context) error {
	out := map[string]func(echo.Context) error{}
	v := reflect.ValueOf(h)
	for i := 0; i < v.NumMethod(); i++ {
		name := v.Type().Method(i).Name
		if !strings.HasSuffix(name, "Handler") || strings.Contains(name, "Webhook") {
			continue
		}
		fn, ok := v.Method(i).Interface().(func(echo.Context) error)
		if ok {
			out[name] = fn
		}
	}
	return out
}

func TestBillingGuardStopsWithoutCompanyOrUser(t *testing.T) {
	cases := []struct {
		name      string
		claims    *pkg.JWTClaims
		company   string
		wantError string
	}{
		{"sem company", &pkg.JWTClaims{UserID: "user-1", Roles: []string{"ROLE_ADMIN"}}, "", "MISSING_TENANT"},
		{"sem userId", &pkg.JWTClaims{Roles: []string{"ROLE_ADMIN"}}, "company-1", `"error":"BAD_REQUEST"`},
	}
	var calls int32
	client := noCallBilling{calls: &calls}
	h := NewBillingHandlers(appbilling.NewBillingPort(client), zap.NewNop()).WithUsers(brPayer())
	handlersByName := billingGuardHandlers(h)
	if len(handlersByName) < 20 {
		t.Fatalf("handlers de billing encontrados: %d", len(handlersByName))
	}
	for _, tc := range cases {
		for name, fn := range handlersByName {
			t.Run(tc.name+"/"+name, func(t *testing.T) {
				before := atomic.LoadInt32(&calls)
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/api/v1/core/billing/x?userId=u-2", strings.NewReader(`{"planCode":"pro","interval":"month","amountCents":1,"reason":"r","q":"a@b.c"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.SetParamNames("id", "code", "gateway", "userId")
				c.SetParamValues("sub-1", "PRO", "asaas", "u-2")
				c.Set("claims", tc.claims)
				if tc.company != "" {
					c.SetRequest(c.Request().WithContext(port.WithCompanyID(c.Request().Context(), tc.company)))
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("o cliente do billing foi chamado: %v", r)
						}
					}()
					_ = fn(c)
				}()
				if got := atomic.LoadInt32(&calls); got != before {
					t.Fatalf("chamadas ao billing: %d", got-before)
				}
				if rec.Code < 400 {
					t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
				}
				if rec.Code == http.StatusBadRequest && !strings.Contains(rec.Body.String(), tc.wantError) && !strings.Contains(rec.Body.String(), "VALIDATION_ERROR") {
					t.Fatalf("erro esperado %s, veio %s", tc.wantError, rec.Body.String())
				}
			})
		}
	}
}
