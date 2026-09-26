package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

// userResponse é o DTO de usuário exposto pela API — separado da entidade de
// domínio para não vazar campos internos.
type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Onboarded   bool   `json:"onboarded"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Onboarded:   u.Onboarded(),
	}
}

// Me devolve o usuário autenticado (a partir do ID no contexto, posto pelo
// middleware de auth).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, err := h.users.GetByID(ctx, userIDFromContext(ctx))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Sessão válida mas usuário sumiu: limpa o cookie.
			h.clearSessionCookie(w)
			writeError(w, http.StatusUnauthorized, "unauthorized", "session no longer valid")
			return
		}
		h.logger.ErrorContext(ctx, "get me failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

type updateMeRequest struct {
	DisplayName string `json:"display_name"`
}

// UpdateMe conclui o onboarding definindo o nome de exibição escolhido.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	u, err := h.users.CompleteOnboarding(ctx, userIDFromContext(ctx), req.DisplayName)
	if err != nil {
		var ve *user.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, "validation_error", ve.Error())
			return
		}
		if errors.Is(err, user.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "session no longer valid")
			return
		}
		h.logger.ErrorContext(ctx, "update me failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}
