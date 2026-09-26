package auth

import (
	"testing"
	"time"
)

func TestTokenManager_IssueAndVerify(t *testing.T) {
	m := NewTokenManager("test-secret", time.Hour)

	token, expires, err := m.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if !expires.After(time.Now()) {
		t.Fatalf("expected expiry in the future, got %v", expires)
	}

	userID, err := m.Verify(token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("userID = %q, want %q", userID, "user-123")
	}
}

func TestTokenManager_Verify(t *testing.T) {
	m := NewTokenManager("test-secret", time.Hour)
	valid, _, _ := m.Issue("user-123")

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{name: "valid token", token: valid, wantErr: false},
		{name: "garbage", token: "not-a-jwt", wantErr: true},
		{name: "empty", token: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.Verify(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify(%q) error = %v, wantErr %v", tt.token, err, tt.wantErr)
			}
		})
	}
}

func TestTokenManager_VerifyRejectsExpired(t *testing.T) {
	m := NewTokenManager("test-secret", -time.Minute) // já expirado
	token, _, err := m.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if _, err := m.Verify(token); err == nil {
		t.Error("expected error verifying expired token, got nil")
	}
}

func TestTokenManager_VerifyRejectsWrongSecret(t *testing.T) {
	issuer := NewTokenManager("secret-a", time.Hour)
	verifier := NewTokenManager("secret-b", time.Hour)

	token, _, _ := issuer.Issue("user-123")
	if _, err := verifier.Verify(token); err == nil {
		t.Error("expected error verifying token signed with a different secret, got nil")
	}
}
