package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

// UserRepo implementa user.Repository sobre o Postgres.
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

const userColumns = `id, google_sub, email, email_verified, display_name, avatar_url, created_at, updated_at, onboarded_at`

func (r *UserRepo) FindByGoogleSub(ctx context.Context, sub string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE google_sub = $1`, sub)
	return scanUser(row)
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (`+userColumns+`)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		u.ID, u.GoogleSub, u.Email, u.EmailVerified, u.DisplayName,
		u.AvatarURL, u.CreatedAt, u.UpdatedAt, u.OnboardedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *user.User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users
		 SET email = $2, email_verified = $3, display_name = $4,
		     avatar_url = $5, updated_at = $6, onboarded_at = $7
		 WHERE id = $1`,
		u.ID, u.Email, u.EmailVerified, u.DisplayName,
		u.AvatarURL, u.UpdatedAt, u.OnboardedAt)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return user.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*user.User, error) {
	var u user.User
	err := row.Scan(&u.ID, &u.GoogleSub, &u.Email, &u.EmailVerified,
		&u.DisplayName, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.OnboardedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
