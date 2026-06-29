package auth

import "time"

type Role string

const (
	RoleUser       Role = "user"
	RoleSupervisor Role = "supervisor"
	RoleAdmin      Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleSupervisor, RoleAdmin:
		return true
	default:
		return false
	}
}

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Claims struct {
	UserID string `json:"sub"`
	Name   string `json:"name"`
	Role   Role   `json:"role"`
}
