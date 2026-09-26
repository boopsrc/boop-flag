package user

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

// fakeRepo é um Repository em memória para testar o Service sem banco.
type fakeRepo struct {
	byID  map[string]*User
	bySub map[string]*User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*User{}, bySub: map[string]*User{}}
}

func (f *fakeRepo) FindByGoogleSub(_ context.Context, sub string) (*User, error) {
	if u, ok := f.bySub[sub]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) FindByID(_ context.Context, id string) (*User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) Create(_ context.Context, u *User) error {
	f.byID[u.ID] = u
	f.bySub[u.GoogleSub] = u
	return nil
}

func (f *fakeRepo) Update(_ context.Context, u *User) error {
	if _, ok := f.byID[u.ID]; !ok {
		return ErrNotFound
	}
	f.byID[u.ID] = u
	f.bySub[u.GoogleSub] = u
	return nil
}

func newTestService() *Service {
	return NewService(newFakeRepo(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestService_UpsertFromGoogle_CreatesThenReuses(t *testing.T) {
	svc := newTestService()
	profile := GoogleProfile{Sub: "sub-1", Email: "a@b.com", EmailVerified: true, Name: "Ana"}

	u1, created, err := svc.UpsertFromGoogle(context.Background(), profile)
	if err != nil {
		t.Fatalf("first upsert error: %v", err)
	}
	if !created {
		t.Error("expected created=true on first upsert")
	}
	if u1.DisplayName != "Ana" {
		t.Errorf("display name = %q, want %q", u1.DisplayName, "Ana")
	}
	if u1.Onboarded() {
		t.Error("new user should not be onboarded yet")
	}

	u2, created, err := svc.UpsertFromGoogle(context.Background(), profile)
	if err != nil {
		t.Fatalf("second upsert error: %v", err)
	}
	if created {
		t.Error("expected created=false on second upsert")
	}
	if u2.ID != u1.ID {
		t.Errorf("expected same user id, got %q and %q", u1.ID, u2.ID)
	}
}

func TestService_UpsertFromGoogle_DefaultNameFromEmail(t *testing.T) {
	svc := newTestService()
	u, _, err := svc.UpsertFromGoogle(context.Background(),
		GoogleProfile{Sub: "sub-2", Email: "carlos@example.com", EmailVerified: true})
	if err != nil {
		t.Fatalf("upsert error: %v", err)
	}
	if u.DisplayName != "carlos" {
		t.Errorf("display name = %q, want %q", u.DisplayName, "carlos")
	}
}

func TestService_CompleteOnboarding(t *testing.T) {
	svc := newTestService()
	created, _, _ := svc.UpsertFromGoogle(context.Background(),
		GoogleProfile{Sub: "sub-3", Email: "d@e.com", EmailVerified: true, Name: "Dora"})

	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{name: "valid", displayName: "Dora Boop", wantErr: false},
		{name: "empty", displayName: "   ", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.CompleteOnboarding(context.Background(), created.ID, tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CompleteOnboarding error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("expected ValidationError, got %T", err)
				}
				return
			}
			if !u.Onboarded() {
				t.Error("expected user to be onboarded after completing")
			}
			if u.DisplayName != "Dora Boop" {
				t.Errorf("display name = %q, want %q", u.DisplayName, "Dora Boop")
			}
		})
	}
}
