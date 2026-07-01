package admin

import (
	"time"

	"optistock/internal/auth"
)

type UserRow struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      auth.Role `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserInput struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     auth.Role `json:"role"`
}

type UpdateUserInput struct {
	Name     *string    `json:"name"`
	Email    *string    `json:"email"`
	Password *string    `json:"password"`
	Role     *auth.Role `json:"role"`
}

type ConfigEntry struct {
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var allowedConfigKeys = map[string]bool{
	"NEAR_EXPIRY_DAYS_DEFAULT":  true,
	"LOW_STOCK_ALERT_ENABLED":   true,
	"NEAR_EXPIRY_ALERT_ENABLED": true,
}
