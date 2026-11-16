package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	authDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/auth"
	userDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/user"
	domainclient "github.com/keepguard/bff-core/internal/application/port"
)

func sameCompany(a, b string) bool {
	a, b = strings.ToLower(strings.TrimSpace(a)), strings.ToLower(strings.TrimSpace(b))
	return a != "" && a == b
}

// patchPersonDocumentScoped (W11): the target user must belong to the request company, and the call goes with
// that company's service token. Refused before any HTTP call otherwise.
func (c *userClient) patchPersonDocumentScoped(ctx context.Context, reqURL, userID, cpf, companyID, correlationID string) (userDto.MSUserResponseDTO, error) {
	if !sameCompany(domainclient.TargetCompanyIDFromContext(ctx), companyID) {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("%w: target user company differs from request company", ErrCrossCompanyCall)
	}
	bearer, err := c.scopedBearer(ctx, companyID)
	if err != nil {
		return userDto.MSUserResponseDTO{}, err
	}
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(map[string]string{"cpf": cpf}).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyID).
		SetHeader("Authorization", bearer).
		SetHeader("Content-Type", "application/json").
		Patch(reqURL)
	if err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao comunicar com user service: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return userDto.MSUserResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "user service")
	}
	var user userDto.MSUserResponseDTO
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}
	if user.CompanyID != "" && !sameCompany(user.CompanyID, companyID) {
		return userDto.MSUserResponseDTO{}, fmt.Errorf("%w: response of another company", ErrCrossCompanyCall)
	}
	return user, nil
}

// getByEmailScoped (W11): the lookup company must be the request company; ms-auth is called with that company's
// service token and the profile enrichment is discarded if ms-user answers for another company.
// No unauthenticated ms-user fallback.
func (c *userClient) getByEmailScoped(ctx context.Context, email, companyID, correlationID string) (authDto.UserByEmailResponseDTO, error) {
	if reqCompany := domainclient.CompanyIDFromContext(ctx); reqCompany != "" && !sameCompany(reqCompany, companyID) {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("%w: lookup company differs from request company", ErrCrossCompanyCall)
	}
	if strings.TrimSpace(companyID) == "" {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("%w: company is required", ErrCrossCompanyCall)
	}
	bearer, err := c.scopedBearer(ctx, companyID)
	if err != nil {
		return authDto.UserByEmailResponseDTO{}, err
	}
	authURL := fmt.Sprintf("%s/api/v1/users/email/%s", strings.TrimRight(c.config.Services.Auth.BaseURL, "/"), url.PathEscape(email))
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyID).
		SetHeader("Authorization", bearer).
		SetHeader("Content-Type", "application/json").
		Get(authURL)
	if err != nil {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao comunicar com auth service: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return authDto.UserByEmailResponseDTO{}, MapHTTPError(resp.StatusCode(), resp.Body(), "auth service")
	}
	var authUser authDto.UserByEmailResponseDTO
	if err := json.Unmarshal(resp.Body(), &authUser); err != nil {
		return authDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}
	if codeUser := strings.TrimSpace(authUser.CodeUser); codeUser != "" {
		userURL := fmt.Sprintf("%s/internal/v1/users/code/%s", strings.TrimRight(c.config.Services.User.BaseURL, "/"), url.PathEscape(codeUser))
		userResp, userErr := c.httpClient.R().
			SetContext(ctx).
			SetHeader("X-Correlation-ID", correlationID).
			SetHeader("X-Company-Id", companyID).
			SetHeader("Content-Type", "application/json").
			Get(userURL)
		if userErr == nil && userResp.StatusCode() == http.StatusOK {
			var msUser userDto.MSUserResponseDTO
			if json.Unmarshal(userResp.Body(), &msUser) == nil && sameCompany(msUser.CompanyID, companyID) {
				if msUser.PersonProfile != nil && strings.TrimSpace(msUser.PersonProfile.FullName) != "" {
					authUser.Username = strings.TrimSpace(msUser.PersonProfile.FullName)
				} else if strings.TrimSpace(msUser.DisplayHandle) != "" {
					authUser.Username = strings.TrimSpace(msUser.DisplayHandle)
				}
			}
		}
	}
	return authUser, nil
}
