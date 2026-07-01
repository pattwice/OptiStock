package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"optistock/internal/auth"
	"optistock/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context) ([]UserRow, error) {
	return s.repo.ListUsers(ctx)
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (*UserRow, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password

	if name == "" || email == "" || password == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "name, email, and password are required")
	}
	if len(password) < 8 {
		return nil, apperror.WithMessage(apperror.ErrValidation, "password must be at least 8 characters")
	}
	if !input.Role.IsValid() {
		return nil, apperror.WithMessage(apperror.ErrValidation, "invalid role")
	}

	existing, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "email already in use")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return s.repo.InsertUser(ctx, name, email, hash, string(input.Role))
}

func (s *Service) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*UserRow, error) {
	existing, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.ErrNotFound
	}

	var name, email, passwordHash, role *string

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, apperror.WithMessage(apperror.ErrValidation, "name cannot be empty")
		}
		name = &trimmed
	}
	if input.Email != nil {
		trimmed := strings.TrimSpace(strings.ToLower(*input.Email))
		if trimmed == "" {
			return nil, apperror.WithMessage(apperror.ErrValidation, "email cannot be empty")
		}
		otherID, err := s.repo.FindUserByEmail(ctx, trimmed)
		if err != nil {
			return nil, err
		}
		if otherID != "" && otherID != id {
			return nil, apperror.WithMessage(apperror.ErrValidation, "email already in use")
		}
		email = &trimmed
	}
	if input.Password != nil {
		if len(*input.Password) < 8 {
			return nil, apperror.WithMessage(apperror.ErrValidation, "password must be at least 8 characters")
		}
		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}
	if input.Role != nil {
		if !input.Role.IsValid() {
			return nil, apperror.WithMessage(apperror.ErrValidation, "invalid role")
		}
		r := string(*input.Role)
		role = &r
	}

	return s.repo.UpdateUser(ctx, id, name, email, passwordHash, role)
}

func (s *Service) ListConfig(ctx context.Context) ([]ConfigEntry, error) {
	return s.repo.ListConfig(ctx)
}

func (s *Service) PatchConfig(ctx context.Context, updates map[string]string) ([]ConfigEntry, error) {
	if len(updates) == 0 {
		return nil, apperror.WithMessage(apperror.ErrValidation, "at least one config key is required")
	}

	var out []ConfigEntry
	for key, value := range updates {
		if !allowedConfigKeys[key] {
			return nil, apperror.WithMessage(apperror.ErrValidation, fmt.Sprintf("config key %q is not editable", key))
		}
		if err := validateConfigValue(key, value); err != nil {
			return nil, err
		}
		entry, err := s.repo.UpsertConfig(ctx, key, value)
		if err != nil {
			return nil, err
		}
		out = append(out, *entry)
	}
	return out, nil
}

func validateConfigValue(key, value string) error {
	switch key {
	case "NEAR_EXPIRY_DAYS_DEFAULT":
		days, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || days < 1 || days > 365 {
			return apperror.WithMessage(apperror.ErrValidation, "NEAR_EXPIRY_DAYS_DEFAULT must be between 1 and 365")
		}
	case "LOW_STOCK_ALERT_ENABLED", "NEAR_EXPIRY_ALERT_ENABLED":
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized != "TRUE" && normalized != "FALSE" {
			return apperror.WithMessage(apperror.ErrValidation, key+" must be TRUE or FALSE")
		}
	}
	return nil
}
