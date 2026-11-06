package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	httpserver "github.com/keepguard/bff-core/internal/adapters/in/http"
	middlewarePkg "github.com/keepguard/bff-core/internal/adapters/in/http/middleware"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/infrastructure/config"
)

const testServiceKey = "valor-da-chave-do-bff-core-so-de-teste"

type sessionsStub struct {
	entries []port.ActiveSession
}

func (s sessionsStub) IsRevoked(context.Context, string) (bool, error) { return false, nil }
func (s sessionsStub) Entries(context.Context, string) ([]port.ActiveSession, error) {
	return s.entries, nil
}

type accessStub struct{ access port.UserAccess }

func (a accessStub) GetUserAccess(context.Context, string, string, string, string) (port.UserAccess, error) {
	return a.access, nil
}

func listedGuard() func(httpserver.Server) httpserver.Server {
	guard := middlewarePkg.NewBillingMoneyGuard([]string{testCompany}, true,
		sessionsStub{entries: []port.ActiveSession{{SID: "sid-1", CodeUser: testUser, CompanyID: testCompany}}},
		accessStub{access: port.UserAccess{Active: true, Roles: []string{"ROLE_ADMIN"}}}, nil)
	return func(server httpserver.Server) httpserver.Server { return server.WithBillingMoneyGuard(guard) }
}

func operatorToken(t *testing.T) string {
	return signToken(t, testUser, []string{"ROLE_ADMIN"}, []string{"billing:read"}, map[string]any{"sid": "sid-1"})
}

// moneyRoutes é a tabela da §3 e do K7 na fase C3. Rota nova de escrita de Operador entra aqui.
var moneyRoutes = []struct {
	name   string
	method string
	path   string
	body   string
}{
	{"POST credits", http.MethodPost, "/credits", `{"userId":"` + testOther + `","amountCents":1000,"reason":"teste"}`},
	{"POST credits reverse", http.MethodPost, "/credits/" + testSubID + "/reverse", `{"reason":"teste"}`},
	{"POST invoices refunds", http.MethodPost, "/invoices/" + testSubID + "/refunds", `{"amountCents":2000,"method":"GATEWAY","reason":"teste"}`},
	{"POST entitlements release-restriction", http.MethodPost, "/entitlements/" + testOther + "/release-restriction", `{"reason":"teste"}`},
	{"PUT billing-settings", http.MethodPut, "/billing-settings", `{"graceDays":7,"reason":"teste"}`},
	{"POST test-accounts", http.MethodPost, "/test-accounts", `{"userId":"` + testOther + `","note":"teste"}`},
	{"DELETE test-accounts", http.MethodDelete, "/test-accounts?userId=" + testOther, ""},
	{"POST billing-reviews resolve", http.MethodPost, "/billing-reviews/" + testSubID + "/resolve", `{"action":"DISMISS","note":"teste"}`},
}

const coreBase = "/api/v1/core/billing"

// K7: lista vazia recusa cada rota, mesmo com ROLE_ADMIN válido e com a chave configurada,
// sem chegar ao ms-billing. Vale sem guard e com guard de lista vazia.
func TestK7EmptyListRefusesEveryOperatorWrite(t *testing.T) {
	upstream := newFakeBilling(t)
	guard := middlewarePkg.NewBillingMoneyGuard(nil, true, sessionsStub{}, accessStub{}, nil)
	for _, variant := range []struct {
		name string
		base string
	}{
		{"sem guard", startBillingRouter(t, baseConfigWithKey(upstream.server.URL))},
		{"lista vazia", startBillingRouter(t, baseConfigWithKey(upstream.server.URL), func(server httpserver.Server) httpserver.Server { return server.WithBillingMoneyGuard(guard) })},
	} {
		for _, route := range moneyRoutes {
			t.Run(variant.name+"/"+route.name, func(t *testing.T) {
				status, _, body := doRequest(t, variant.base, goldenCase{method: route.method, path: coreBase + route.path, body: route.body, token: operatorToken(t), headers: map[string]string{"Idempotency-Key": "k1"}})
				if status != http.StatusForbidden || !strings.Contains(body, "CORE_WRITE_DISABLED") {
					t.Fatalf("status %d body %s", status, body)
				}
				if calls := upstream.drain(); len(calls) != 0 {
					t.Fatalf("não pode chegar ao ms-billing: %+v", calls)
				}
			})
		}
	}
}

func baseConfigWithKey(url string) *config.Config {
	cfg := baseConfig(url)
	cfg.BillingCore.ServiceKey = testServiceKey
	return cfg
}

// K2: billing:write e assinante recusados em dinheiro, antes da lista, por rota.
func TestK2NonOperatorRefusedOnEveryRoute(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL), listedGuard())
	tokens := map[string]string{
		"billing:write":           signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:write", "billing:read"}, map[string]any{"sid": "sid-1"}),
		"ROLE_ORGANIZATION_ADMIN": signToken(t, testUser, []string{"ROLE_ORGANIZATION_ADMIN"}, []string{"billing:read"}, map[string]any{"sid": "sid-1"}),
		"assinante":               signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read", "billing:self-write"}, map[string]any{"sid": "sid-1"}),
	}
	for tokenName, token := range tokens {
		for _, route := range moneyRoutes {
			t.Run(tokenName+"/"+route.name, func(t *testing.T) {
				status, _, body := doRequest(t, base, goldenCase{method: route.method, path: coreBase + route.path, body: route.body, token: token})
				if status != http.StatusForbidden || !strings.Contains(body, "FORBIDDEN") {
					t.Fatalf("status %d body %s", status, body)
				}
			})
		}
	}
	if calls := upstream.drain(); len(calls) != 0 {
		t.Fatalf("não pode chegar ao ms-billing: %d chamadas", len(calls))
	}
}

// K7.1: com a lista preenchida e sessão real, passa com credencial, operador, chave de
// idempotência e IP real. O corpo vai com o número intacto.
func TestK71ListedCompanyForwardsWithCredential(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL), listedGuard())
	for _, route := range moneyRoutes {
		t.Run(route.name, func(t *testing.T) {
			status, _, body := doRequest(t, base, goldenCase{method: route.method, path: coreBase + route.path, body: route.body, token: operatorToken(t),
				headers: map[string]string{"Idempotency-Key": "chave-1", "X-Forwarded-For": "6.6.6.6, 200.200.200.200", "X-Billing-Service-Key": "forjada", "X-Operator-Id": testOther}})
			if status != http.StatusOK {
				t.Fatalf("status %d body %s", status, body)
			}
			calls := upstream.drain()
			if len(calls) != 1 {
				t.Fatalf("chamadas %d", len(calls))
			}
			call := calls[0]
			wantURI := "/api/v1/billing" + route.path
			if call.Method != route.method || call.URI != wantURI {
				t.Fatalf("%s %s, esperado %s", call.Method, call.URI, wantURI)
			}
			assertHeader(t, call, "X-Billing-Service-Key", testServiceKey)
			assertHeader(t, call, "X-Operator-Id", testUser)
			assertHeader(t, call, "Idempotency-Key", "chave-1")
			assertHeader(t, call, "X-Company-Id", testCompany)
			assertHeader(t, call, "X-Real-Ip", "200.200.200.200")
			if route.body != "" {
				var sent, original map[string]any
				if err := json.Unmarshal([]byte(call.Body), &sent); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(route.body), &original); err != nil {
					t.Fatal(err)
				}
				for key, value := range original {
					if sent[key] != value {
						t.Fatalf("campo %s: %v, enviado %v", key, value, sent[key])
					}
				}
			}
		})
	}
}

func assertHeader(t *testing.T, call upstreamCall, name, want string) {
	t.Helper()
	if got := call.Headers[name]; len(got) != 1 || got[0] != want {
		t.Fatalf("cabeçalho %s = %v, esperado %s", name, got, want)
	}
}

// K7.1: a chave nunca liga company nem muda piso e tetos, em nenhuma grafia.
func TestK71NeverTurnsCompanyOn(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL), listedGuard())
	for _, body := range []string{
		`{"financialMode":"ON","reason":"x"}`,
		`{"FINANCIALMODE":"on","reason":"x"}`,
		`{"financialMode":"OFF","FinancialMode":"ON","reason":"x"}`,
		`{"minChargeCents":500,"reason":"x"}`,
		`{"creditCapBp":5000,"reason":"x"}`,
		`{"dailyCreditCapCents":1,"reason":"x"}`,
		`{"dailyRefundCapCents":1,"reason":"x"}`,
		`{"daily_refund_cap_cents":1,"reason":"x"}`,
	} {
		t.Run(body, func(t *testing.T) {
			status, _, response := doRequest(t, base, goldenCase{method: http.MethodPut, path: coreBase + "/billing-settings", body: body, token: operatorToken(t)})
			if status != http.StatusForbidden || !strings.Contains(response, "CORE_KEY_NOT_ALLOWED") {
				t.Fatalf("status %d body %s", status, response)
			}
			if calls := upstream.drain(); len(calls) != 0 {
				t.Fatal("não pode chegar ao ms-billing")
			}
		})
	}
	status, _, _ := doRequest(t, base, goldenCase{method: http.MethodPut, path: coreBase + "/billing-settings", body: `{"financialMode":"DRAINING","reason":"x"}`, token: operatorToken(t)})
	if status != http.StatusOK {
		t.Fatalf("DRAINING é decisão do billing: %d", status)
	}
	upstream.drain()
}

// K7.3 e erros do §3: teto, alerta e configuração moram no ms-billing, por qualquer caminho.
// O bff-core repassa status e código.
func TestUpstreamErrorsKeepStatusAndCode(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL), listedGuard())
	cases := []struct {
		name   string
		status int
		code   string
	}{
		{"teto diario", http.StatusTooManyRequests, "DAILY_CAP_EXCEEDED"},
		{"chave repetida", http.StatusConflict, "IDEMPOTENCY_KEY_REUSED"},
		{"sem chave", http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED"},
		{"em curso", http.StatusConflict, "REQUEST_IN_PROGRESS"},
		{"lista do billing", http.StatusForbidden, "CORE_WRITE_DISABLED"},
		{"chave do bff-core", http.StatusUnauthorized, "SERVICE_KEY_INVALID"},
		{"recurso desligado", http.StatusNotFound, "FEATURE_DISABLED"},
		{"proprio operador", http.StatusForbidden, "SELF_ACTION_FORBIDDEN"},
		{"configuracao indisponivel", http.StatusServiceUnavailable, "SETTINGS_UNAVAILABLE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			upstream.replies["POST /api/v1/billing/credits"] = fakeReply{tc.status, `{"timestamp":"x","status":0,"error":"` + tc.code + `","message":"mensagem"}`}
			status, _, body := doRequest(t, base, goldenCase{method: http.MethodPost, path: coreBase + "/credits", body: `{"amountCents":1}`, token: operatorToken(t), headers: map[string]string{"Idempotency-Key": "k"}})
			if status != tc.status || !strings.Contains(body, tc.code) {
				t.Fatalf("status %d body %s", status, body)
			}
			if calls := upstream.drain(); len(calls) != 1 {
				t.Fatalf("escrita não repete sozinha: %d chamadas", len(calls))
			}
		})
	}
	delete(upstream.replies, "POST /api/v1/billing/credits")
}

// Cartão no corpo também recusa na rota de Operador, sem chegar ao ms-billing.
func TestMoneyRouteRefusesCardData(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL), listedGuard())
	status, _, body := doRequest(t, base, goldenCase{method: http.MethodPost, path: coreBase + "/credits", body: `{"amountCents":1,"creditCard":{"number":"4111"}}`, token: operatorToken(t)})
	if status != http.StatusUnprocessableEntity || !strings.Contains(body, "PAYMENT_METHOD_UNSUPPORTED") {
		t.Fatalf("status %d body %s", status, body)
	}
	if calls := upstream.drain(); len(calls) != 0 {
		t.Fatal("não pode chegar ao ms-billing")
	}
}

// Rotas de leitura e de assinante: seguem com JWT, sem credencial nem trava de escrita.
func TestReadAndSubscriberRoutesForward(t *testing.T) {
	upstream := newFakeBilling(t)
	upstream.replies["DELETE /api/v1/billing/subscriptions/"+testSubID+"/card"] = fakeReply{http.StatusNoContent, ""}
	base := startBillingRouter(t, baseConfigWithKey(upstream.server.URL))
	subscriber := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read", "billing:self-write"}, map[string]any{"sid": "sid-1"})
	operator := operatorToken(t)
	cases := []struct {
		name    string
		method  string
		path    string
		token   string
		status  int
		wantURI string
	}{
		{"extrato proprio ignora userId", http.MethodGet, "/credits?page=1&size=5&userId=" + testOther, subscriber, http.StatusOK, "/api/v1/billing/credits?page=1&size=5"},
		{"extrato de operador", http.MethodGet, "/credits?userId=" + testOther, operator, http.StatusOK, "/api/v1/billing/credits?userId=" + testOther},
		{"remover cartao", http.MethodDelete, "/subscriptions/" + testSubID + "/card", subscriber, http.StatusNoContent, "/api/v1/billing/subscriptions/" + testSubID + "/card"},
		{"configuracao de operador", http.MethodGet, "/billing-settings", operator, http.StatusOK, "/api/v1/billing/billing-settings"},
		{"revisoes de operador", http.MethodGet, "/billing-reviews?status=OPEN", operator, http.StatusOK, "/api/v1/billing/billing-reviews?status=OPEN"},
		{"previa de estorno", http.MethodGet, "/invoices/" + testSubID + "/refund-preview?amountCents=100&method=GATEWAY", operator, http.StatusOK, "/api/v1/billing/invoices/" + testSubID + "/refund-preview?amountCents=100&method=GATEWAY"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, body := doRequest(t, base, goldenCase{method: tc.method, path: coreBase + tc.path, token: tc.token})
			if status != tc.status {
				t.Fatalf("status %d body %s", status, body)
			}
			calls := upstream.drain()
			if len(calls) != 1 || calls[0].URI != tc.wantURI {
				t.Fatalf("chamadas %+v", calls)
			}
			for _, header := range []string{"X-Billing-Service-Key", "X-Operator-Id", "X-Real-Ip"} {
				if _, present := calls[0].Headers[header]; present {
					t.Fatalf("%s não vai em leitura", header)
				}
			}
		})
	}
	for _, path := range []string{"/billing-settings", "/billing-reviews", "/invoices/" + testSubID + "/refund-preview"} {
		status, _, _ := doRequest(t, base, goldenCase{method: http.MethodGet, path: coreBase + path, token: subscriber})
		if status != http.StatusForbidden {
			t.Fatalf("%s de assinante: %d", path, status)
		}
	}
	if calls := upstream.drain(); len(calls) != 0 {
		t.Fatalf("assinante não chega ao ms-billing em rota de operador: %+v", calls)
	}
}

// Idempotency-Key e campos novos chegam ao ms-billing nas trocas (§3, §6.7, §7.1). K8, do lado
// do bff-core: o JWT renovado vai intacto, e quem confere billing:self-write é o billing.
func TestSubscriberWritesForwardIdempotencyAndNewFields(t *testing.T) {
	upstream := newFakeBilling(t)
	upstream.replies["POST /api/v1/billing/subscriptions"] = fakeReply{http.StatusCreated, `{"status":"pending_gateway"}`}
	upstream.replies["DELETE /api/v1/billing/subscriptions/"+testSubID+"/card"] = fakeReply{http.StatusNoContent, ""}
	base := startBillingRouter(t, baseConfig(upstream.server.URL))
	subscriber := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read", "billing:self-write"}, map[string]any{"sid": "sid-1"})
	sub := "/subscriptions/" + testSubID
	cases := []struct {
		name     string
		method   string
		path     string
		body     string
		wantBody []string
	}{
		{"assinar com aceite", http.MethodPost, "/subscriptions", `{"planCode":"PRO","interval":"month","paymentMethod":"credit_card","cardDebitConsent":{"textVersion":"v1"}}`, []string{`"cardDebitConsent":{"textVersion":"v1"}`}},
		{"trocar com quoteId", http.MethodPost, sub + "/change-plan", `{"planCode":"AGENCIA","interval":"month","quoteId":"q-1","cardDebitConsent":{"textVersion":"v2"}}`, []string{`"quoteId":"q-1"`, `"cardDebitConsent":{"textVersion":"v2"}`}},
		{"cancelar", http.MethodPost, sub + "/cancel", "", nil},
		{"reativar com aceite", http.MethodPost, sub + "/reactivate", `{"cardDebitConsent":{"textVersion":"v3"}}`, []string{`"cardDebitConsent":{"textVersion":"v3"}`}},
		{"desfazer troca", http.MethodDelete, sub + "/scheduled-change", "", nil},
		{"remover cartao", http.MethodDelete, sub + "/card", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, body := doRequest(t, base, goldenCase{method: tc.method, path: coreBase + tc.path, body: tc.body, token: subscriber, headers: map[string]string{"Idempotency-Key": "abc-1"}})
			if status >= 300 {
				t.Fatalf("status %d body %s", status, body)
			}
			calls := upstream.drain()
			if len(calls) != 1 {
				t.Fatalf("chamadas %d", len(calls))
			}
			assertHeader(t, calls[0], "Idempotency-Key", "abc-1")
			if got := calls[0].Headers["Authorization"]; len(got) != 1 || got[0] != "Bearer "+subscriber {
				t.Fatalf("JWT precisa ir intacto: %v", got)
			}
			if _, present := calls[0].Headers["X-Billing-Service-Key"]; present {
				t.Fatal("credencial não vai em rota de assinante")
			}
			for _, want := range tc.wantBody {
				if !strings.Contains(calls[0].Body, want) {
					t.Fatalf("corpo %s sem %s", calls[0].Body, want)
				}
			}
		})
	}
}

// R1.4 (§9, §11): IP e aparelho do corpo e X-Public-IP são ignorados.
func TestGlobalExceptionIPFromConnection(t *testing.T) {
	upstream := newFakeBilling(t)
	upstream.replies["POST /api/v1/billing/subscriptions"] = fakeReply{http.StatusCreated, `{"status":"active"}`}
	base := startBillingRouter(t, baseConfig(upstream.server.URL))
	subscriber := signToken(t, testUser, []string{"ROLE_USER"}, []string{"billing:read"}, nil)
	post := func(t *testing.T, body string, headers map[string]string) map[string]any {
		t.Helper()
		status, _, response := doRequest(t, base, goldenCase{method: http.MethodPost, path: coreBase + "/subscriptions", body: body, token: subscriber, headers: headers})
		if status != http.StatusCreated {
			t.Fatalf("status %d body %s", status, response)
		}
		calls := upstream.drain()
		if len(calls) != 1 {
			t.Fatalf("chamadas %d", len(calls))
		}
		var sent map[string]any
		if err := json.Unmarshal([]byte(calls[0].Body), &sent); err != nil {
			t.Fatal(err)
		}
		return sent
	}
	t.Run("corpo e X-Public-IP ignorados", func(t *testing.T) {
		sent := post(t, `{"planCode":"PRO","interval":"month","paymentMethod":"pix","ipAddress":"9.9.9.9","deviceId":"dev-corpo","remoteIp":"9.9.9.9"}`,
			map[string]string{"X-Public-IP": "8.8.8.8", "X-Device-Id": "dev-header", "X-Real-IP": "200.200.200.1"})
		if sent["ipAddress"] != "200.200.200.1" || sent["deviceId"] != "dev-header" {
			t.Fatalf("enviado %v", sent)
		}
		if _, present := sent["remoteIp"]; present {
			t.Fatalf("remoteIp do corpo não vai: %v", sent)
		}
	})
	t.Run("aparelho sem cabecalho vem do token", func(t *testing.T) {
		sent := post(t, `{"planCode":"PRO","interval":"month","paymentMethod":"pix","deviceId":"dev-corpo"}`, nil)
		if sent["deviceId"] != "dev-claim" {
			t.Fatalf("enviado %v", sent)
		}
	})
	t.Run("remoteIp nao e o primeiro do X-Forwarded-For", func(t *testing.T) {
		sent := post(t, `{"planCode":"PRO","interval":"month","paymentMethod":"credit_card","creditCardToken":"tok","remoteIp":"9.9.9.9"}`,
			map[string]string{"X-Forwarded-For": "6.6.6.6, 200.200.200.2"})
		if sent["remoteIp"] != "200.200.200.2" || sent["ipAddress"] != "200.200.200.2" {
			t.Fatalf("enviado %v", sent)
		}
		if sent["creditCardToken"] != "tok" {
			t.Fatalf("token segue aceito neste caminho: %v", sent)
		}
	})
}

// S1 e S2, do lado do bff-core: o corpo do webhook vai ao ms-billing byte a byte. A assinatura
// da Stripe e a máscara do payload são conferidas lá, sobre este corpo.
func TestWebhooksForwardRawBytes(t *testing.T) {
	upstream := newFakeBilling(t)
	base := startBillingRouter(t, baseConfig(upstream.server.URL))
	raw := `{ "id":"evt_1", "payment":{"value":97.00,"creditCardToken":"tok"},  "x":"a&b" }`
	for _, tc := range []struct{ path, header, value string }{
		{"/webhooks/asaas", "asaas-access-token", "tok-do-webhook"},
		{"/webhooks/stripe", "stripe-signature", "t=1,v1=assinatura"},
	} {
		req, err := http.NewRequest(http.MethodPost, base+coreBase+tc.path, strings.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(tc.header, tc.value)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d", tc.path, resp.StatusCode)
		}
		calls := upstream.drain()
		if len(calls) != 1 || calls[0].Body != raw {
			t.Fatalf("%s: corpo alterado: %+v", tc.path, calls)
		}
		assertHeader(t, calls[0], http.CanonicalHeaderKey(tc.header), tc.value)
	}
}
