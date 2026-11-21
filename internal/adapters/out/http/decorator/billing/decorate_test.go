package billing

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go.uber.org/zap"

	"github.com/keepguard/bff-core/internal/application/port"
)

// countingBilling falha com erro de rede (que o observe considera repetível) e conta as chamadas.
type countingBilling struct {
	port.BillingClient
	changes, reads, cancels int
}

var errNetwork = errors.New("dial tcp: i/o timeout")

func (c *countingBilling) ChangePlan(context.Context, port.BillingScope, string, any) (json.RawMessage, error) {
	c.changes++
	return nil, errNetwork
}

func (c *countingBilling) CancelSubscription(context.Context, port.BillingScope, string, bool) (json.RawMessage, error) {
	c.cancels++
	return nil, errNetwork
}

func (c *countingBilling) GetEntitlement(context.Context, port.BillingScope) (json.RawMessage, error) {
	c.reads++
	return nil, errNetwork
}

// Escrita de cobrança não se repete sozinha (revisão de 21/11/2026): um timeout pode ter sido aplicado no ms-billing.
func TestBillingWritesAreNotRetried(t *testing.T) {
	inner := &countingBilling{}
	d := New(inner, zap.NewNop(), nil, "test")
	ctx := context.Background()

	_, _ = d.ChangePlan(ctx, port.BillingScope{}, "sub-1", map[string]any{})
	_, _ = d.CancelSubscription(ctx, port.BillingScope{}, "sub-1", false)
	_, _ = d.GetEntitlement(ctx, port.BillingScope{})

	if inner.changes != 1 {
		t.Fatalf("change-plan chamado %d vezes, esperado 1", inner.changes)
	}
	if inner.cancels != 1 {
		t.Fatalf("cancel chamado %d vezes, esperado 1", inner.cancels)
	}
	if inner.reads != 2 {
		t.Fatalf("leitura deve continuar com nova tentativa: %d chamadas, esperado 2", inner.reads)
	}
}
