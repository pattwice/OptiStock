package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`, email)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`, id)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.NewString(), userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) FindRefreshToken(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash)
	if err := row.Scan(&userID, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, fmt.Errorf("refresh token not found")
		}
		return "", time.Time{}, err
	}
	return userID, expiresAt, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *Repository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	return err
}
