package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/boopsrc/boop-flag/backend/internal/auth"
	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

const (
	sessionCookieName = "boop_session"
	stateCookieName   = "boop_oauth_state"
)

// Handler agrupa as dependências dos handlers HTTP. Handlers ficam finos: só
// parseiam entrada, chamam o Service/auth e serializam a saída.
type Handler struct {
	users    *user.Service
	tokens   *auth.TokenManager
	states   *auth.StateManager
	google   *auth.GoogleProvider
	logger   *slog.Logger
	frontend string
	ttl      time.Duration
	devMode  bool
	secure   bool
}

type HandlerConfig struct {
	Users       *user.Service
	Tokens      *auth.TokenManager
	States      *auth.StateManager
	Google      *auth.GoogleProvider
	Logger      *slog.Logger
	FrontendURL string
	SessionTTL  time.Duration
	DevMode     bool
	// Secure controla a flag Secure do cookie; deve ser true em produção
	// (HTTPS) e false em dev sobre HTTP.
	Secure bool
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		users:    cfg.Users,
		tokens:   cfg.Tokens,
		states:   cfg.States,
		google:   cfg.Google,
		logger:   cfg.Logger,
		frontend: cfg.FrontendURL,
		ttl:      cfg.SessionTTL,
		devMode:  cfg.DevMode,
		secure:   cfg.Secure,
	}
}

// setSessionCookie grava o JWT de sessão num cookie HttpOnly.
func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// issueSession emite o JWT para o usuário e grava o cookie de sessão.
func (h *Handler) issueSession(w http.ResponseWriter, userID string) error {
	token, expires, err := h.tokens.Issue(userID)
	if err != nil {
		return err
	}
	h.setSessionCookie(w, token, expires)
	return nil
}
