package saga

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type repoFake struct {
	estados map[string]*State
	erroGet error
	erroSet error
	apagou  []string
}

func novoRepo() *repoFake {
	return &repoFake{estados: map[string]*State{}}
}

func (r *repoFake) Save(ctx context.Context, state *State) error {
	if r.erroSet != nil {
		return r.erroSet
	}
	r.estados[state.SagaID] = state
	return nil
}

func (r *repoFake) Get(ctx context.Context, sagaID string) (*State, error) {
	if r.erroGet != nil {
		return nil, r.erroGet
	}
	s, ok := r.estados[sagaID]
	if !ok {
		return nil, errors.New("não encontrado")
	}
	return s, nil
}

func (r *repoFake) Delete(ctx context.Context, sagaID string) error {
	r.apagou = append(r.apagou, sagaID)
	delete(r.estados, sagaID)
	return nil
}

func gerente(repo StateRepository) *StateManager {
	return NewStateManager(repo, zap.NewNop())
}

func TestNewState(t *testing.T) {
	s := NewState("saga-1")

	if s.SagaID != "saga-1" {
		t.Errorf("SagaID = %q", s.SagaID)
	}
	// saga nasce rodando, no passo zero
	if !s.IsRunning() {
		t.Errorf("status inicial = %v, quer running", s.Status)
	}
	if s.CurrentStep != 0 {
		t.Errorf("CurrentStep = %d, quer 0", s.CurrentStep)
	}
	if s.Data == nil {
		t.Error("Data nil; quem escreve nele faria panic")
	}
	if s.CompletedSteps == nil {
		t.Error("CompletedSteps nil")
	}
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		t.Error("timestamps não preenchidos")
	}
}

func TestState_Transicoes(t *testing.T) {
	casos := []struct {
		status SagaStatus
		checa  func(*State) bool
		nome   string
	}{
		{StatusRunning, (*State).IsRunning, "running"},
		{StatusCompleted, (*State).IsCompleted, "completed"},
		{StatusFailed, (*State).IsFailed, "failed"},
		{StatusCompensated, (*State).IsCompensated, "compensated"},
		{StatusCompensating, (*State).IsCompensating, "compensating"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			s := NewState("x")
			s.SetStatus(c.status)

			if !c.checa(s) {
				t.Errorf("Is%s() = false após SetStatus(%v)", c.nome, c.status)
			}
			// os outros predicados têm que ser falsos
			for _, outro := range casos {
				if outro.status != c.status && outro.checa(s) {
					t.Errorf("Is%s() também devolveu true", outro.nome)
				}
			}
		})
	}
}

func TestState_AddCompletedStep(t *testing.T) {
	s := NewState("x")

	s.AddCompletedStep("passo-1")
	s.AddCompletedStep("passo-2")

	if len(s.CompletedSteps) != 2 {
		t.Fatalf("CompletedSteps = %v, quer 2 itens", s.CompletedSteps)
	}
	// a ordem importa: é ela que diz o que compensar, e em que sequência
	if s.CompletedSteps[0] != "passo-1" || s.CompletedSteps[1] != "passo-2" {
		t.Errorf("ordem errada: %v", s.CompletedSteps)
	}
}

func TestStateManager_SalvaECarrega(t *testing.T) {
	repo := novoRepo()
	sm := gerente(repo)
	ctx := context.Background()

	s := NewState("saga-1")
	s.AddCompletedStep("passo-1")

	if err := sm.SaveState(ctx, s); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	lido, err := sm.LoadState(ctx, "saga-1")
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if lido.SagaID != "saga-1" || len(lido.CompletedSteps) != 1 {
		t.Errorf("estado carregado diverge: %+v", lido)
	}
}

func TestStateManager_LoadInexistenteFalha(t *testing.T) {
	sm := gerente(novoRepo())

	if _, err := sm.LoadState(context.Background(), "não-existe"); err == nil {
		t.Fatal("carregou saga inexistente")
	}
}

func TestStateManager_DeleteState(t *testing.T) {
	repo := novoRepo()
	sm := gerente(repo)
	ctx := context.Background()

	_ = sm.SaveState(ctx, NewState("saga-1"))
	if err := sm.DeleteState(ctx, "saga-1"); err != nil {
		t.Fatalf("DeleteState: %v", err)
	}
	if len(repo.apagou) != 1 || repo.apagou[0] != "saga-1" {
		t.Errorf("apagou = %v", repo.apagou)
	}
}

// Retomar uma saga já finalizada executaria de novo efeitos que já
// aconteceram — cobrar duas vezes, mandar dois e-mails.
func TestStateManager_ResumeRecusaSagaFinalizada(t *testing.T) {
	casos := []struct {
		nome       string
		status     SagaStatus
		querNoErro string
	}{
		{"completada", StatusCompleted, "completada"},
		{"falhada", StatusFailed, "falhou"},
		{"compensada", StatusCompensated, "compensada"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			repo := novoRepo()
			sm := gerente(repo)
			ctx := context.Background()

			s := NewState("saga-1")
			s.SetStatus(c.status)
			_ = sm.SaveState(ctx, s)

			saga, err := sm.ResumeSaga(ctx, "saga-1", []Step{})
			if err == nil {
				t.Fatal("retomou saga finalizada")
			}
			if saga != nil {
				t.Error("devolveu saga junto com o erro")
			}
			if !strings.Contains(err.Error(), c.querNoErro) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}
}

func TestStateManager_ResumeSagaRodando(t *testing.T) {
	repo := novoRepo()
	sm := gerente(repo)
	ctx := context.Background()

	s := NewState("saga-1")
	s.AddCompletedStep("passo-1")
	s.CurrentStep = 1
	_ = sm.SaveState(ctx, s)

	steps := []Step{NewStepBuilder("passo-2").WithExecute(execNoop).Build()}
	saga, err := sm.ResumeSaga(ctx, "saga-1", steps)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// a retomada tem que preservar onde parou
	if saga.State.CurrentStep != 1 {
		t.Errorf("CurrentStep = %d, quer 1", saga.State.CurrentStep)
	}
	if len(saga.Steps) != 1 {
		t.Errorf("Steps = %d, quer 1", len(saga.Steps))
	}
}

func TestStateManager_GetSagaStatus(t *testing.T) {
	repo := novoRepo()
	sm := gerente(repo)
	ctx := context.Background()

	s := NewState("saga-1")
	s.SetStatus(StatusCompensating)
	_ = sm.SaveState(ctx, s)

	got, err := sm.GetSagaStatus(ctx, "saga-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != StatusCompensating {
		t.Errorf("status = %v, quer %v", got, StatusCompensating)
	}
}

func TestStateManager_MarshalUnmarshal(t *testing.T) {
	sm := gerente(novoRepo())

	original := NewState("saga-1")
	original.AddCompletedStep("passo-1")
	original.Data["chave"] = "valor"

	raw, err := sm.MarshalState(original)
	if err != nil {
		t.Fatalf("MarshalState: %v", err)
	}

	lido, err := sm.UnmarshalState(raw)
	if err != nil {
		t.Fatalf("UnmarshalState: %v", err)
	}
	if lido.SagaID != original.SagaID {
		t.Errorf("SagaID = %q", lido.SagaID)
	}
	if len(lido.CompletedSteps) != 1 {
		t.Errorf("CompletedSteps = %v", lido.CompletedSteps)
	}
	if lido.Data["chave"] != "valor" {
		t.Errorf("Data = %v", lido.Data)
	}
}

func TestStateManager_UnmarshalInvalidoFalha(t *testing.T) {
	sm := gerente(novoRepo())

	if _, err := sm.UnmarshalState([]byte("não é json")); err == nil {
		t.Fatal("aceitou JSON inválido")
	}
}

func TestStateManager_ValidateState(t *testing.T) {
	sm := gerente(novoRepo())

	t.Run("estado válido passa", func(t *testing.T) {
		if err := sm.ValidateState(NewState("saga-1")); err != nil {
			t.Errorf("recusou estado válido: %v", err)
		}
	})

	// Validar preenche o que estava nil: quem chamar depois faz range e
	// escrita sem checar.
	t.Run("preenche Data e CompletedSteps nil", func(t *testing.T) {
		s := &State{SagaID: "x", Status: StatusRunning}
		if err := sm.ValidateState(s); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if s.Data == nil {
			t.Error("Data seguiu nil")
		}
		if s.CompletedSteps == nil {
			t.Error("CompletedSteps seguiu nil")
		}
	})

	casos := []struct {
		nome       string
		state      *State
		querNoErro string
	}{
		{"sem SagaID", &State{Status: StatusRunning}, "SagaID"},
		{"sem status", &State{SagaID: "x"}, "Status"},
		{"status inventado", &State{SagaID: "x", Status: "VOANDO"}, "inválido"},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			err := sm.ValidateState(c.state)
			if err == nil {
				t.Fatal("aceitou estado inválido")
			}
			if !strings.Contains(err.Error(), c.querNoErro) {
				t.Errorf("erro = %q, esperava conter %q", err, c.querNoErro)
			}
		})
	}
}

// O estado é serializado para o repositório — os campos precisam sobreviver
// à ida e volta com os nomes que o JSON declara.
func TestState_SerializacaoPreservaCampos(t *testing.T) {
	s := NewState("saga-1")
	s.SetStatus(StatusCompensating)
	s.AddCompletedStep("passo-1")
	s.CurrentStep = 3

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var lido State
	if err := json.Unmarshal(raw, &lido); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if lido.SagaID != "saga-1" || lido.Status != StatusCompensating || lido.CurrentStep != 3 {
		t.Errorf("estado perdido na serialização: %+v", lido)
	}
}
