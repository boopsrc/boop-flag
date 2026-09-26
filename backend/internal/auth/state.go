package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StateManager gera e valida o parâmetro `state` do fluxo OAuth, protegendo
// contra CSRF. O state é um nonce aleatório com timestamp, assinado com HMAC —
// assim conseguimos validá-lo sem guardar estado no servidor. Ele é enviado ao
// Google e também gravado num cookie; no callback os dois precisam bater.
type StateManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewStateManager(secret string) *StateManager {
	return &StateManager{
		secret: []byte(secret),
		ttl:    10 * time.Minute,
		now:    time.Now,
	}
}

// New gera um valor de state assinado no formato "nonce.timestamp.assinatura".
func (m *StateManager) New() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	ts := strconv.FormatInt(m.now().Unix(), 10)
	payload := nonce + "." + ts
	sig := m.sign(payload)
	return payload + "." + sig, nil
}

// Validate confere assinatura e expiração do state.
func (m *StateManager) Validate(state string) error {
	parts := strings.Split(state, ".")
	if len(parts) != 3 {
		return fmt.Errorf("malformed state")
	}
	payload := parts[0] + "." + parts[1]
	expectedSig := m.sign(payload)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return fmt.Errorf("invalid state signature")
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid state timestamp: %w", err)
	}
	if m.now().Sub(time.Unix(ts, 0)) > m.ttl {
		return fmt.Errorf("state expired")
	}
	return nil
}

func (m *StateManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
