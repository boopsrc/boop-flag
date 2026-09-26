package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boopsrc/boop-flag/backend/internal/auth"
	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

type stubRepo struct {
	users map[string]*user.User
}

func (s *stubRepo) FindByGoogleSub(context.Context, string) (*user.User, error) {
	return nil, user.ErrNotFound
}
func (s *stubRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, user.ErrNotFound
}
func (s *stubRepo) Create(_ context.Context, u *user.User) error { s.users[u.ID] = u; return nil }
func (s *stubRepo) Update(_ context.Context, u *user.User) error { s.users[u.ID] = u; return nil }

func newTestServer(t *testing.T, repo *stubRepo) (http.Handler, *auth.TokenManager) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := user.NewService(repo, logger)
	tokens := auth.NewTokenManager("test-secret", time.Hour)
	states := auth.NewStateManager("test-secret")
	h := NewHandler(HandlerConfig{
		Users: svc, Tokens: tokens, States: states, Google: nil,
		Logger: logger, FrontendURL: "http://localhost:5173",
		SessionTTL: time.Hour, DevMode: true, Secure: false,
	})
	return NewRouter(h, tokens, logger, "http://localhost:5173"), tokens
}

func TestMe_Unauthenticated(t *testing.T) {
	srv, _ := newTestServer(t, &stubRepo{users: map[string]*user.User{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMe_Authenticated(t *testing.T) {
	repo := &stubRepo{users: map[string]*user.User{
		"user-1": {ID: "user-1", Email: "a@b.com", DisplayName: "Ana"},
	}}
	srv, tokens := newTestServer(t, repo)

	token, expires, _ := tokens.Issue("user-1")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token, Expires: expires})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp userResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Email != "a@b.com" || resp.DisplayName != "Ana" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestHealthz(t *testing.T) {
	srv, _ := newTestServer(t, &stubRepo{users: map[string]*user.User{}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
