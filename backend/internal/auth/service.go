package auth

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
	"optistock/internal/config"
	"optistock/pkg/apperror"
)

const refreshCookieName = "optistock_refresh"

type Service struct {
	repo    *Repository
	tokens  *TokenManager
	refresh time.Duration
}

func NewService(repo *Repository, tokens *TokenManager, cfg *config.Config) *Service {
	return &Service{
		repo:    repo,
		tokens:  tokens,
		refresh: cfg.RefreshTokenTTL,
	}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	User        UserPublic  `json:"user"`
	AccessToken string      `json:"access_token"`
	ExpiresIn   int64       `json:"expires_in"`
	RefreshToken string     `json:"-"`
}

type UserPublic struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	accessToken, expiresIn, err := s.tokens.IssueAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.refresh)); err != nil {
		return nil, err
	}

	return &LoginResult{
		User: UserPublic{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:  accessToken,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	if refreshToken == "" {
		return nil, apperror.ErrInvalidToken
	}

	hash := HashToken(refreshToken)
	userID, expiresAt, err := s.repo.FindRefreshToken(ctx, hash)
	if err != nil || time.Now().After(expiresAt) {
		return nil, apperror.ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, apperror.ErrInvalidToken
	}

	if err := s.repo.DeleteRefreshToken(ctx, hash); err != nil {
		return nil, err
	}

	accessToken, expiresIn, err := s.tokens.IssueAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefresh, newHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveRefreshToken(ctx, user.ID, newHash, time.Now().Add(s.refresh)); err != nil {
		return nil, err
	}

	return &LoginResult{
		User: UserPublic{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:  accessToken,
		ExpiresIn:    expiresIn,
		RefreshToken: newRefresh,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repo.DeleteRefreshToken(ctx, HashToken(refreshToken))
}

func (s *Service) Me(ctx context.Context, userID string) (*UserPublic, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}
	return &UserPublic{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func RefreshCookieName() string {
	return refreshCookieName
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
