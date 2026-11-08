package handlers

import (
	"net/http/httptest"
	"testing"
)

// O IP do cadastro é o que o servidor vê (X-Real-IP do nginx), não o que o navegador põe no corpo.
func TestIPDoCadastro(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/v1/register/init", nil)
	r.Header.Set("X-Real-IP", "189.45.12.8")
	if ip := ipDoCadastro(r, "1.2.3.4"); ip != "189.45.12.8" {
		t.Errorf("ip = %q; o do servidor tem de vencer o do corpo", ip)
	}

	r = httptest.NewRequest("POST", "/api/v1/register/init", nil)
	r.RemoteAddr = ""
	if ip := ipDoCadastro(r, " 1.2.3.4 "); ip != "1.2.3.4" {
		t.Errorf("ip = %q; sem IP do servidor, vale o do corpo", ip)
	}
}
