package register

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

var errReplayCorrupt = errors.New("resultado de replay ilegível")

// confirmMaterial junta o que só quem fez o cadastro conhece: a sessão (UUID) e os códigos. Dele saem a chave do
// Redis (hash) e a chave de cifra do resultado. Quem lê o Redis sem esse material não abre o token de login.
func confirmMaterial(command appdto.RegisterConfirmCommand) string {
	return strings.Join([]string{
		command.RegistrationSessionId,
		command.Token,
		command.EmailToken,
		command.SmsToken,
		command.WhatsAppToken,
	}, "\x1f")
}

// confirmFingerprint é a parte pública da chave (sha256 do material, em hex).
func confirmFingerprint(material string) string {
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])
}

// deriveReplayKey separa a chave de cifra da chave do Redis (domínio próprio, mesma fonte).
func deriveReplayKey(material string) []byte {
	sum := sha256.Sum256([]byte("kg:bff-core:register:result-key\x1f" + material))
	return sum[:]
}

// sealReplay cifra o resultado com AES-256-GCM (nonce aleatório na frente).
func sealReplay(material string, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveReplayKey(material))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

// openReplay decifra o resultado; material errado ou dado adulterado devolve erro.
func openReplay(material string, sealed []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveReplayKey(material))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(sealed) < gcm.NonceSize() {
		return nil, errReplayCorrupt
	}
	plain, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return nil, errReplayCorrupt
	}
	return plain, nil
}
