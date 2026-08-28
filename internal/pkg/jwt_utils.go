package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTClaims representa as claims do JWT
type JWTClaims struct {
	CodeUser  string   `json:"codeUser"`
	Sub       string   `json:"sub"`
	Username  string   `json:"username"`
	TenantId  string   `json:"tenant_id"`
	CompanyID string   `json:"companyId"`
	UserID    string   `json:"userId"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
}

// ExtractCodeUserFromToken extrai o codeUser do token JWT sem validar a assinatura
// Esta função apenas decodifica o payload do token para obter informações
// A validação do token é feita pelo ms-auth
func ExtractCodeUserFromToken(token string) (string, error) {
	// Remove o prefixo "Bearer " se existir
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Split do token em suas partes (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("token JWT inválido: formato incorreto")
	}

	// Decodifica o payload (segunda parte)
	payload := parts[1]

	// Adiciona padding se necessário para Base64
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	// Decodifica de Base64
	decodedPayload, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar payload do token: %w", err)
	}

	// Parse do JSON
	var claims JWTClaims
	if err := json.Unmarshal(decodedPayload, &claims); err != nil {
		return "", fmt.Errorf("erro ao fazer parse das claims do token: %w", err)
	}

	// Tenta primeiro codeUser, depois sub, depois username
	if claims.CodeUser != "" {
		return claims.CodeUser, nil
	}
	if claims.Sub != "" {
		return claims.Sub, nil
	}
	if claims.Username != "" {
		return claims.Username, nil
	}

	return "", fmt.Errorf("codeUser não encontrado no token")
}

// ExtractAllClaims extrai todas as claims do token JWT sem validar a assinatura
func ExtractAllClaims(token string) (*JWTClaims, error) {
	// Remove o prefixo "Bearer " se existir
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Split do token em suas partes (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token JWT inválido: formato incorreto")
	}

	// Decodifica o payload (segunda parte)
	payload := parts[1]

	// Adiciona padding se necessário para Base64
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	// Decodifica de Base64
	decodedPayload, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar payload do token: %w", err)
	}

	// Parse do JSON
	var claims JWTClaims
	if err := json.Unmarshal(decodedPayload, &claims); err != nil {
		return nil, fmt.Errorf("erro ao fazer parse das claims do token: %w", err)
	}

	return &claims, nil
}
