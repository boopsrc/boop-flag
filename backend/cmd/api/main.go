package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/boopsrc/boop-flag/backend/internal/auth"
	"github.com/boopsrc/boop-flag/backend/internal/config"
	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
	"github.com/boopsrc/boop-flag/backend/internal/httpapi"
	"github.com/boopsrc/boop-flag/backend/internal/storage/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// SIGINT/SIGTERM cancelam o contexto raiz, disparando o graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("connected to database")

	userRepo := postgres.NewUserRepo(pool)
	userSvc := user.NewService(userRepo, logger)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.SessionTTL)
	states := auth.NewStateManager(cfg.JWTSecret)

	var google *auth.GoogleProvider
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		google = auth.NewGoogleProvider(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	}

	handler := httpapi.NewHandler(httpapi.HandlerConfig{
		Users:       userSvc,
		Tokens:      tokens,
		States:      states,
		Google:      google,
		Logger:      logger,
		FrontendURL: cfg.FrontendURL,
		SessionTTL:  cfg.SessionTTL,
		DevMode:     cfg.AuthDevMode,
		Secure:      strings.HasPrefix(cfg.FrontendURL, "https://"),
	})

	router := httpapi.NewRouter(handler, tokens, logger, cfg.FrontendURL)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", srv.Addr, "dev_mode", cfg.AuthDevMode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
