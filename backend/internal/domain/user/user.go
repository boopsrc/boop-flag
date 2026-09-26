package user

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound indica que o usuário procurado não existe.
var ErrNotFound = errors.New("user not found")

// User é a entidade de domínio. Representa uma conta criada a partir de um
// login Google (ou de um login mock em modo dev).
type User struct {
	ID            string
	GoogleSub     string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// OnboardedAt fica nil até o usuário completar o onboarding (escolher o
	// nome de exibição). É o que distingue um "cadastro" recém-criado de uma
	// conta já estabelecida.
	OnboardedAt *time.Time
}

// Onboarded diz se o usuário já concluiu o onboarding.
func (u *User) Onboarded() bool {
	return u.OnboardedAt != nil
}

// Repository abstrai a persistência de usuários. A interface vive no pacote que
// a consome (domínio), e é implementada pela camada de storage — isso permite
// testar o Service com um repositório fake, sem banco.
type Repository interface {
	FindByGoogleSub(ctx context.Context, sub string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
}
