// Package redisstore lê as chaves de sessão que o ms-auth grava no Redis.
package redisstore

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/keepguard/bff-core/internal/application/port"
	"github.com/redis/go-redis/v9"
)

const (
	sessionRevokedPrefix = "session:revoked:"
	sidTokensPrefix      = "refreshtoken:sid:"
	refreshTokenPrefix   = "refreshtoken:"
)

var errRedisUnavailable = errors.New("redis indisponível")

type activeSessionReader struct {
	client *redis.Client
}

// NewActiveSessionReader monta o leitor. Sem cliente, toda leitura falha: quem usa
// recusa com 503 (falha fechada).
func NewActiveSessionReader(client *redis.Client) port.ActiveSessionReader {
	return &activeSessionReader{client: client}
}

func (r *activeSessionReader) IsRevoked(ctx context.Context, sid string) (bool, error) {
	if r.client == nil {
		return false, errRedisUnavailable
	}
	count, err := r.client.Exists(ctx, sessionRevokedPrefix+sid).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Entries segue o índice refreshtoken:sid:<sid> até as entradas refreshtoken:<hash>. Índice
// ausente devolve lista vazia. Entrada expirada ou ilegível não conta como sessão.
func (r *activeSessionReader) Entries(ctx context.Context, sid string) ([]port.ActiveSession, error) {
	if r.client == nil {
		return nil, errRedisUnavailable
	}
	hashes, err := r.client.SMembers(ctx, sidTokensPrefix+sid).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]port.ActiveSession, 0, len(hashes))
	for _, hash := range hashes {
		raw, err := r.client.Get(ctx, refreshTokenPrefix+hash).Bytes()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var entry refreshTokenEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		entries = append(entries, port.ActiveSession{SID: entry.SID, CodeUser: entry.CodeUser, CompanyID: entry.CompanyID, Revoked: entry.Revoked})
	}
	return entries, nil
}

// refreshTokenEntry são os campos usados do JSON do ms-auth (RefreshTokenModel).
type refreshTokenEntry struct {
	SID       string `json:"sid"`
	CodeUser  string `json:"codeUser"`
	CompanyID string `json:"companyId"`
	Revoked   bool   `json:"revoked"`
}
