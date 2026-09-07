package register

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

// RegisterInitUseCase define o caso de uso de inicialização de registro
type RegisterInitUseCase interface {
	Execute(ctx context.Context, command appdto.RegisterInitCommand) (appdto.RegisterInitViewDTO, error)
}

// RegisterConfirmUseCase define o caso de uso de confirmação de registro
type RegisterConfirmUseCase interface {
	Execute(ctx context.Context, command appdto.RegisterConfirmCommand) (appdto.RegisterConfirmViewDTO, error)
}
