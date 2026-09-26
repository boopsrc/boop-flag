package auth

import (
	"testing"
	"time"
)

func TestStateManager_NewAndValidate(t *testing.T) {
	m := NewStateManager("test-secret")

	state, err := m.New()
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if err := m.Validate(state); err != nil {
		t.Errorf("Validate returned error for fresh state: %v", err)
	}
}

func TestStateManager_ValidateRejectsBad(t *testing.T) {
	m := NewStateManager("test-secret")
	valid, _ := m.New()

	tests := []struct {
		name  string
		state string
	}{
		{name: "malformed", state: "abc"},
		{name: "empty", state: ""},
		{name: "tampered signature", state: valid[:len(valid)-3] + "xxx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Validate(tt.state); err == nil {
				t.Errorf("Validate(%q) = nil, want error", tt.state)
			}
		})
	}
}

func TestStateManager_ValidateRejectsExpired(t *testing.T) {
	m := NewStateManager("test-secret")
	// Força um relógio no passado ao gerar, deixando o state fora da janela.
	m.now = func() time.Time { return time.Now().Add(-time.Hour) }
	state, _ := m.New()

	m.now = time.Now
	if err := m.Validate(state); err == nil {
		t.Error("expected error validating expired state, got nil")
	}
}
