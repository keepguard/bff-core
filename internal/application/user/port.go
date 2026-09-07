package user

import (
	"context"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type UserPort interface {
	GetMe(ctx context.Context, query appdto.GetMeQuery) (appdto.MeProfileViewDTO, error)
}
