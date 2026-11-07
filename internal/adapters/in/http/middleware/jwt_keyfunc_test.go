package http

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	libauth "github.com/keepguard/lib-go-common/auth"
)

func tokenEdDSA(t *testing.T, kid string, priv ed25519.PrivateKey) string {
	t.Helper()
	tk := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{"sub": "u", "exp": time.Now().Add(time.Minute).Unix()})
	tk.Header["kid"] = kid
	s, err := tk.SignedString(priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestValidateTokenLocal_EdDSAViaJWKSeHS256EmConvivencia(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	keys := libauth.StaticKeys{"k1": pub}
	secret := "segredo-de-teste-com-mais-de-32-bytes!!"

	hs, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "u", "exp": time.Now().Add(time.Minute).Unix()}).SignedString([]byte(secret))
	ed := tokenEdDSA(t, "k1", priv)
	edKidErrado := tokenEdDSA(t, "k9", priv)

	convivencia := &JWTMiddleware{jwtSecret: secret, keys: keys}
	if _, err := convivencia.validateTokenLocal(hs, ""); err != nil {
		t.Fatalf("HS256 na convivência: %v", err)
	}
	if _, err := convivencia.validateTokenLocal(ed, ""); err != nil {
		t.Fatalf("EdDSA via JWKS: %v", err)
	}
	if _, err := convivencia.validateTokenLocal(edKidErrado, ""); err == nil {
		t.Fatal("kid desconhecido não pode passar")
	}

	final := &JWTMiddleware{jwtSecret: "", keys: keys} // fim da migração: sem segredo no pod
	if _, err := final.validateTokenLocal(hs, ""); err == nil {
		t.Fatal("sem segredo, HS256 não pode passar")
	}
	if _, err := final.validateTokenLocal(ed, ""); err != nil {
		t.Fatalf("EdDSA no estado final: %v", err)
	}
	soHS := &JWTMiddleware{jwtSecret: secret}
	if _, err := soHS.validateTokenLocal(ed, ""); err == nil {
		t.Fatal("sem JWKS, EdDSA não passa")
	}
}
