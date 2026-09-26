package httpapi

import (
	"net/http"
	"time"

	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

// GoogleLogin inicia o fluxo OAuth: gera um state assinado, grava-o num cookie
// curto e redireciona o usuário para a tela de consentimento do Google.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "google login is not configured")
		return
	}

	state, err := h.states.New()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "generate state failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   int((10 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.google.AuthCodeURL(state), http.StatusFound)
}

// GoogleCallback recebe o retorno do Google, valida o state, troca o código
// pelo perfil, cria/atualiza o usuário e emite a sessão.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		writeError(w, http.StatusServiceUnavailable, "oauth_disabled", "google login is not configured")
		return
	}
	ctx := r.Context()

	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(stateCookieName)
	if err != nil || cookie.Value == "" || cookie.Value != state {
		writeError(w, http.StatusBadRequest, "invalid_state", "invalid oauth state")
		return
	}
	if err := h.states.Validate(state); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_state", "invalid oauth state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing_code", "missing authorization code")
		return
	}

	profile, err := h.google.Exchange(ctx, code)
	if err != nil {
		h.logger.ErrorContext(ctx, "google exchange failed", "error", err)
		writeError(w, http.StatusBadGateway, "oauth_exchange_failed", "could not complete google sign-in")
		return
	}
	if !profile.EmailVerified {
		writeError(w, http.StatusForbidden, "email_unverified", "google email is not verified")
		return
	}

	h.finishLogin(w, r, profile)
}

// DevLogin cria/loga um usuário fake sem passar pelo Google. Só existe quando
// AUTH_DEV_MODE está ligado; serve para desenvolvimento e demonstração.
func (h *Handler) DevLogin(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		email = "amiguinho@boop.dev"
	}

	profile := user.GoogleProfile{
		Sub:           "dev-" + email,
		Email:         email,
		EmailVerified: true,
		Name:          "",
		Picture:       "",
	}
	h.finishLogin(w, r, profile)
}

// finishLogin faz o upsert do usuário, emite a sessão e redireciona para o
// frontend — para o onboarding se for um cadastro novo, senão para a home.
func (h *Handler) finishLogin(w http.ResponseWriter, r *http.Request, profile user.GoogleProfile) {
	ctx := r.Context()

	u, created, err := h.users.UpsertFromGoogle(ctx, profile)
	if err != nil {
		h.logger.ErrorContext(ctx, "upsert user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	if err := h.issueSession(w, u.ID); err != nil {
		h.logger.ErrorContext(ctx, "issue session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	// Limpa o cookie de state, que já cumpriu seu papel.
	http.SetCookie(w, &http.Cookie{
		Name: stateCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode,
	})

	dest := h.frontend + "/home"
	if created || !u.Onboarded() {
		dest = h.frontend + "/onboarding"
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

// Logout limpa o cookie de sessão.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
