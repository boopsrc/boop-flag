package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GoogleProfile são os dados vindos do provedor OAuth (Google) necessários para
// criar ou atualizar um usuário.
type GoogleProfile struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// Service concentra as regras de negócio de usuário. Handlers HTTP chamam o
// Service; o Service não conhece HTTP nem SQL diretamente.
type Service struct {
	repo   Repository
	logger *slog.Logger
	now    func() time.Time
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger, now: time.Now}
}

// UpsertFromGoogle cria o usuário no primeiro acesso (o "cadastro") ou atualiza
// os dados vindos do Google num acesso posterior. Retorna o usuário e se ele
// acabou de ser criado.
func (s *Service) UpsertFromGoogle(ctx context.Context, p GoogleProfile) (*User, bool, error) {
	if p.Sub == "" {
		return nil, false, fmt.Errorf("google profile without sub")
	}

	existing, err := s.repo.FindByGoogleSub(ctx, p.Sub)
	switch {
	case err == nil:
		existing.Email = p.Email
		existing.EmailVerified = p.EmailVerified
		if existing.AvatarURL == "" {
			existing.AvatarURL = p.Picture
		}
		existing.UpdatedAt = s.now()
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, false, fmt.Errorf("update user %s: %w", existing.ID, err)
		}
		return existing, false, nil

	case errors.Is(err, ErrNotFound):
		now := s.now()
		u := &User{
			ID:            uuid.NewString(),
			GoogleSub:     p.Sub,
			Email:         p.Email,
			EmailVerified: p.EmailVerified,
			DisplayName:   defaultDisplayName(p),
			AvatarURL:     p.Picture,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.repo.Create(ctx, u); err != nil {
			return nil, false, fmt.Errorf("create user: %w", err)
		}
		s.logger.InfoContext(ctx, "user created", "user_id", u.ID, "email", u.Email)
		return u, true, nil

	default:
		return nil, false, fmt.Errorf("find user by google sub: %w", err)
	}
}

// GetByID busca um usuário pelo ID interno.
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	return u, nil
}

// CompleteOnboarding define o nome de exibição escolhido pelo usuário e marca o
// onboarding como concluído.
func (s *Service) CompleteOnboarding(ctx context.Context, id, displayName string) (*User, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, &ValidationError{Field: "display_name", Message: "required"}
	}
	if len([]rune(displayName)) > 60 {
		return nil, &ValidationError{Field: "display_name", Message: "must be at most 60 characters"}
	}

	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	now := s.now()
	u.DisplayName = displayName
	u.UpdatedAt = now
	if u.OnboardedAt == nil {
		u.OnboardedAt = &now
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update user %s: %w", id, err)
	}
	return u, nil
}

func defaultDisplayName(p GoogleProfile) string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return name
	}
	if at := strings.IndexByte(p.Email, '@'); at > 0 {
		return p.Email[:at]
	}
	return "Amiguinho"
}

// ValidationError representa uma falha de validação de entrada com o campo
// específico que a causou, para o handler mapear em um 400 informativo.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
