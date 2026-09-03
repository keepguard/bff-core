package oauthsecret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

const ivLength = 12

func aesGCM(secretBase string) (cipher.AEAD, error) {
	if secretBase == "" {
		return nil, fmt.Errorf("AUTH_CLIENT_SECRET_BASE é obrigatório")
	}
	sum := sha256.Sum256([]byte(secretBase))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar GCM: %w", err)
	}
	return gcm, nil
}

func DecryptAESGCM(secretBase, encrypted string) (string, error) {
	if encrypted == "" {
		return "", fmt.Errorf("secret cifrado vazio")
	}
	packed, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("secret cifrado inválido: %w", err)
	}
	if len(packed) <= ivLength {
		return "", fmt.Errorf("secret cifrado incompleto")
	}
	gcm, err := aesGCM(secretBase)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, packed[:ivLength], packed[ivLength:], nil)
	if err != nil {
		return "", fmt.Errorf("falha ao descriptografar clientSecret: %w", err)
	}
	return string(plain), nil
}

func EncryptAESGCM(secretBase, plain string) (string, error) {
	if plain == "" {
		return "", fmt.Errorf("clientSecret é obrigatório")
	}
	gcm, err := aesGCM(secretBase)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("falha ao gerar IV: %w", err)
	}
	sealed := gcm.Seal(nil, nonce, []byte(plain), nil)
	packed := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(packed), nil
}
