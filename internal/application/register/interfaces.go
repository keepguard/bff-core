package register

import (
	"github.com/keepguard/bff-core/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

// RegisterInitUseCase define o caso de uso de inicialização de registro
type RegisterInitUseCase interface {
	Execute(command appdto.RegisterInitCommand) (dto.RegisterInitResponseDTO, error)
}

// RegisterConfirmUseCase define o caso de uso de confirmação de registro
type RegisterConfirmUseCase interface {
	Execute(command appdto.RegisterConfirmCommand) (dto.RegisterConfirmResponseDTO, error)
}
