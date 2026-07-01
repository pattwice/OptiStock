package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (r *Repository) ListUsers(ctx context.Context) ([]UserRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, email, role, created_at
		FROM users
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (*UserRow, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, role, created_at
		FROM users
		WHERE id = $1
	`, id)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) InsertUser(ctx context.Context, name, email, passwordHash, role string) (*UserRow, error) {
	id := uuid.NewString()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, role, created_at
	`, id, name, email, passwordHash, role)

	var u UserRow
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id string, name, email, passwordHash, role *string) (*UserRow, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	argN := 1

	if name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, *name)
		argN++
	}
	if email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", argN))
		args = append(args, *email)
		argN++
	}
	if passwordHash != nil {
		sets = append(sets, fmt.Sprintf("password_hash = $%d", argN))
		args = append(args, *passwordHash)
		argN++
	}
	if role != nil {
		sets = append(sets, fmt.Sprintf("role = $%d", argN))
		args = append(args, *role)
		argN++
	}
	if len(sets) == 0 {
		return r.FindUserByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, name, email, role, created_at
	`, strings.Join(sets, ", "), argN)

	row := r.pool.QueryRow(ctx, query, args...)
	var u UserRow
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ListConfig(ctx context.Context) ([]ConfigEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT config_key, config_value, updated_at
		FROM e1_system_config
		ORDER BY config_key ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ConfigEntry
	for rows.Next() {
		var e ConfigEntry
		if err := rows.Scan(&e.ConfigKey, &e.ConfigValue, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertConfig(ctx context.Context, key, value string) (*ConfigEntry, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO e1_system_config (config_key, config_value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (config_key) DO UPDATE
		SET config_value = EXCLUDED.config_value, updated_at = NOW()
		RETURNING config_key, config_value, updated_at
	`, key, value)

	var e ConfigEntry
	if err := row.Scan(&e.ConfigKey, &e.ConfigValue, &e.UpdatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
