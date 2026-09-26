package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/boopsrc/boop-flag/backend/internal/auth"
)

// NewRouter monta a árvore de rotas e a cadeia de middlewares. A ordem dos
// middlewares importa: recover primeiro (captura panics de tudo abaixo),
// depois request ID, logging, CORS; a autenticação é aplicada só ao grupo de
// rotas protegidas.
func NewRouter(h *Handler, tokens *auth.TokenManager, logger *slog.Logger, frontendURL string) http.Handler {
	r := chi.NewRouter()

	r.Use(Recover(logger))
	r.Use(RequestID())
	r.Use(Logging(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Get("/google/login", h.GoogleLogin)
			r.Get("/google/callback", h.GoogleCallback)
			r.Get("/dev/login", h.DevLogin)
			r.Post("/logout", h.Logout)
		})

		// Rotas protegidas: exigem sessão válida.
		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(tokens))
			r.Get("/me", h.Me)
			r.Patch("/me", h.UpdateMe)
		})
	})

	return r
}
