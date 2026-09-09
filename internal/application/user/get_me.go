package user

import (
	"context"
	"net/http"
	"strings"
	"time"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/keepguard/bff-core/internal/pkg"
)

type getMeUseCase struct {
	users     port.UserClient
	companies port.CompanyClient
}

func NewUserPort(users port.UserClient, companies port.CompanyClient) UserPort {
	return &getMeUseCase{users: users, companies: companies}
}

func (uc *getMeUseCase) GetMe(ctx context.Context, query appdto.GetMeQuery) (appdto.MeProfileViewDTO, error) {
	company, err := uc.companies.GetByTenantId(ctx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.MeProfileViewDTO{}, err
	}
	if company.ID == "" {
		return appdto.MeProfileViewDTO{}, pkg.NewAppError(
			"COMPANY_NOT_FOUND",
			"Empresa não encontrada para o tenant informado",
			http.StatusNotFound,
		)
	}

	user, err := uc.users.GetUserByCodeUser(
		port.WithCompanyID(ctx, company.ID),
		query.CodeUser,
		query.Token,
		query.TenantID,
		query.CorrelationID,
	)
	if err != nil {
		return appdto.MeProfileViewDTO{}, err
	}
	return toMeProfile(user), nil
}

func toMeProfile(user appdto.MSUserResponseDTO) appdto.MeProfileViewDTO {
	view := appdto.MeProfileViewDTO{
		Email:           user.Email,
		PhoneE164:       user.PhoneE164,
		PreferredLocale: user.PreferredLocale,
		Timezone:        user.Timezone,
		AvatarURL:       user.AvatarURL,
		DisplayHandle:   user.DisplayHandle,
		Type:            user.Type,
		Status:          user.Status,
	}
	if !user.CreatedAt.IsZero() {
		view.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}
	if user.PersonProfile != nil {
		profile := &appdto.MePersonProfileViewDTO{FullName: user.PersonProfile.FullName}
		digits := documentDigits(user.PersonProfile.CPF)
		if len(digits) == 11 {
			profile.HasCpf = true
			profile.CpfLast4 = digits[len(digits)-4:]
		}
		if profile.FullName != "" || profile.HasCpf {
			view.PersonProfile = profile
		}
	}
	return view
}

func documentDigits(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
