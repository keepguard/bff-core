package saga

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Compensação não para no primeiro erro: todas as compensações rodam (ordem inversa) e os erros são acumulados.
func TestCompensate_ContinuaAposErroEAcumulaErros(t *testing.T) {
	errC := errors.New("falha do passo C")
	var order []string

	mk := func(name string, compErr error) InMemoryStep {
		return InMemoryStep{
			Name:       name,
			Execute:    func(context.Context, map[string]interface{}) error { return nil },
			Compensate: func(context.Context, map[string]interface{}) error { order = append(order, name); return compErr },
			MaxRetries: 1,
			Timeout:    time.Second,
		}
	}
	steps := []InMemoryStep{
		mk("A", nil),
		mk("B", nil),
		mk("C", errC),
		{
			Name:       "D",
			Execute:    func(context.Context, map[string]interface{}) error { return errors.New("falha no D") },
			MaxRetries: 1,
		},
	}
	// D falha (sem compensação); A, B e C foram executados e são desfeitos em ordem inversa. C falha ao desfazer.
	exec := NewInMemorySagaExecutor(zap.NewNop())
	err := exec.Execute(context.Background(), InMemorySaga{Name: "t", Steps: steps}, map[string]interface{}{})

	if err == nil {
		t.Fatal("esperava erro")
	}
	if want := []string{"C", "B", "A"}; !equalStrings(order, want) {
		t.Fatalf("ordem de compensação = %v, quer %v (tem que tentar todas)", order, want)
	}
	if !errors.Is(err, errC) {
		t.Fatalf("o erro acumulado deveria conter a falha do C: %v", err)
	}
}

// Quando a compensação de um passo falha, as outras ainda rodam e o erro devolvido carrega todas as falhas.
func TestCompensate_ErrosDeVariosPassosSaoAcumulados(t *testing.T) {
	errB := errors.New("compensação B falhou")
	errA := errors.New("compensação A falhou")
	var ran []string

	steps := []InMemoryStep{
		{Name: "A", Execute: ok, Compensate: func(context.Context, map[string]interface{}) error { ran = append(ran, "A"); return errA }, MaxRetries: 1},
		{Name: "B", Execute: ok, Compensate: func(context.Context, map[string]interface{}) error { ran = append(ran, "B"); return errB }, MaxRetries: 1},
		{Name: "X", Execute: func(context.Context, map[string]interface{}) error { return errors.New("falha no X") }, MaxRetries: 1},
	}
	err := NewInMemorySagaExecutor(zap.NewNop()).Execute(context.Background(), InMemorySaga{Name: "t", Steps: steps}, map[string]interface{}{})

	if !equalStrings(ran, []string{"B", "A"}) {
		t.Fatalf("compensações = %v, quer [B A]", ran)
	}
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("erro acumulado sem as duas falhas: %v", err)
	}
}

// A compensação roda com contexto próprio: o cancelamento do cliente não a interrompe.
func TestCompensate_ContextoNaoEhCanceladoPeloCliente(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var compCtxErr error
	steps := []InMemoryStep{
		{
			Name:       "A",
			Execute:    ok,
			Compensate: func(c context.Context, _ map[string]interface{}) error { compCtxErr = c.Err(); return nil },
			MaxRetries: 1,
		},
		{
			Name: "B",
			Execute: func(context.Context, map[string]interface{}) error {
				cancel() // cliente desconectou no meio da saga
				return errors.New("falha no B")
			},
			MaxRetries: 1,
		},
	}
	if err := NewInMemorySagaExecutor(zap.NewNop()).Execute(ctx, InMemorySaga{Name: "t", Steps: steps}, map[string]interface{}{}); err == nil {
		t.Fatal("esperava erro")
	}
	if compCtxErr != nil {
		t.Fatalf("compensação recebeu contexto cancelado: %v", compCtxErr)
	}
}

// Cada compensação tem prazo de 30 s (o contexto próprio não fica sem limite).
func TestCompensate_TemPrazoDe30s(t *testing.T) {
	var deadline time.Time
	var hasDeadline bool
	steps := []InMemoryStep{
		{
			Name:    "A",
			Execute: ok,
			Compensate: func(c context.Context, _ map[string]interface{}) error {
				deadline, hasDeadline = c.Deadline()
				return nil
			},
			MaxRetries: 1,
			Timeout:    time.Second, // o prazo do passo não limita a compensação
		},
		{Name: "B", Execute: func(context.Context, map[string]interface{}) error { return errors.New("x") }, MaxRetries: 1},
	}
	_ = NewInMemorySagaExecutor(zap.NewNop()).Execute(context.Background(), InMemorySaga{Name: "t", Steps: steps}, map[string]interface{}{})

	if !hasDeadline {
		t.Fatal("compensação sem prazo")
	}
	left := time.Until(deadline)
	if left < 25*time.Second || left > 31*time.Second {
		t.Fatalf("prazo da compensação = %v, quer ~30s", left)
	}
}

func ok(context.Context, map[string]interface{}) error { return nil }

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
