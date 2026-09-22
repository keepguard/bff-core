package saga

import (
	"context"
	"strings"
	"testing"
	"time"
)

func execNoop(ctx context.Context, state *State) error { return nil }

// O builder já entrega retries e timeout preenchidos: um step sem esses
// valores é recusado pelo validador, e ninguém ia lembrar de definir os dois
// em cada step.
func TestNewStepBuilder_TemDefaults(t *testing.T) {
	step := NewStepBuilder("criar-usuario").Build()

	if step.Name != "criar-usuario" {
		t.Errorf("Name = %q", step.Name)
	}
	if step.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, quer 3", step.MaxRetries)
	}
	if step.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, quer 10s", step.Timeout)
	}
}

func TestStepBuilder_Encadeado(t *testing.T) {
	compensou := false

	step := NewStepBuilder("passo").
		WithExecute(execNoop).
		WithCompensate(func(ctx context.Context, state *State) error {
			compensou = true
			return nil
		}).
		WithMaxRetries(7).
		WithTimeout(30 * time.Second).
		Build()

	if step.Execute == nil {
		t.Error("Execute não foi definido")
	}
	if step.Compensate == nil {
		t.Fatal("Compensate não foi definido")
	}
	if step.MaxRetries != 7 {
		t.Errorf("MaxRetries = %d, quer 7", step.MaxRetries)
	}
	if step.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, quer 30s", step.Timeout)
	}

	if err := step.Compensate(context.Background(), &State{}); err != nil {
		t.Fatalf("compensação: %v", err)
	}
	if !compensou {
		t.Error("a função de compensação registrada não foi a chamada")
	}
}

func TestStepValidator_Validate(t *testing.T) {
	casos := []struct {
		nome       string
		step       Step
		querNoErro string
	}{
		{
			nome:       "sem nome",
			step:       Step{Execute: execNoop, MaxRetries: 1, Timeout: time.Second},
			querNoErro: "nome",
		},
		{
			nome:       "sem função de execução",
			step:       Step{Name: "x", MaxRetries: 1, Timeout: time.Second},
			querNoErro: "execução",
		},
		{
			nome:       "retries zero",
			step:       Step{Name: "x", Execute: execNoop, MaxRetries: 0, Timeout: time.Second},
			querNoErro: "tentativas",
		},
		{
			nome:       "retries negativo",
			step:       Step{Name: "x", Execute: execNoop, MaxRetries: -1, Timeout: time.Second},
			querNoErro: "tentativas",
		},
		{
			nome:       "timeout zero",
			step:       Step{Name: "x", Execute: execNoop, MaxRetries: 1, Timeout: 0},
			querNoErro: "timeout",
		},
	}

	sv := NewStepValidator()

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			err := sv.Validate(c.step)
			if err == nil {
				t.Fatal("aceitou step inválido")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(c.querNoErro)) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}

	t.Run("step válido passa", func(t *testing.T) {
		valido := NewStepBuilder("ok").WithExecute(execNoop).Build()
		if err := sv.Validate(valido); err != nil {
			t.Errorf("recusou step válido: %v", err)
		}
	})
}

func TestStepValidator_ValidateSaga(t *testing.T) {
	sv := NewStepValidator()
	stepOK := NewStepBuilder("ok").WithExecute(execNoop).Build()

	t.Run("saga nil", func(t *testing.T) {
		if err := sv.ValidateSaga(nil); err == nil {
			t.Fatal("aceitou saga nil")
		}
	})

	t.Run("estado nil", func(t *testing.T) {
		err := sv.ValidateSaga(&Saga{State: nil, Steps: []Step{stepOK}})
		if err == nil {
			t.Fatal("aceitou saga sem estado")
		}
		if !strings.Contains(err.Error(), "estado") {
			t.Errorf("erro = %q, esperava falar de estado", err)
		}
	})

	// Saga sem step não compensa nada e não executa nada: é erro de montagem.
	t.Run("sem steps", func(t *testing.T) {
		err := sv.ValidateSaga(&Saga{State: &State{}, Steps: []Step{}})
		if err == nil {
			t.Fatal("aceitou saga sem steps")
		}
		if !strings.Contains(err.Error(), "pelo menos um step") {
			t.Errorf("erro = %q", err)
		}
	})

	// O erro tem que dizer QUAL step está inválido, senão numa saga de 8
	// passos não se sabe onde mexer.
	t.Run("step inválido identifica a posição e o nome", func(t *testing.T) {
		ruim := Step{Name: "passo-ruim", Execute: execNoop, MaxRetries: 0, Timeout: time.Second}
		err := sv.ValidateSaga(&Saga{State: &State{}, Steps: []Step{stepOK, ruim}})
		if err == nil {
			t.Fatal("aceitou saga com step inválido")
		}
		if !strings.Contains(err.Error(), "step 1") {
			t.Errorf("erro = %q, esperava apontar 'step 1'", err)
		}
		if !strings.Contains(err.Error(), "passo-ruim") {
			t.Errorf("erro = %q, esperava citar o nome do step", err)
		}
	})

	t.Run("saga válida passa", func(t *testing.T) {
		s := &Saga{State: &State{SagaID: "s1"}, Steps: []Step{stepOK}}
		if err := sv.ValidateSaga(s); err != nil {
			t.Errorf("recusou saga válida: %v", err)
		}
	})
}
