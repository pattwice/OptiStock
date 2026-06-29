package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"optistock/internal/auth"
	"optistock/pkg/apperror"
	"optistock/pkg/response"
)

func Register(app *fiber.App, origins string) {
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
}

func JWTAuth(tokens *auth.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return response.Fail(c, apperror.ErrUnauthorized)
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokens.ParseAccessToken(token)
		if err != nil {
			return response.Fail(c, apperror.ErrInvalidToken)
		}
		c.Locals("userID", claims.UserID)
		c.Locals("userName", claims.Name)
		c.Locals("userRole", string(claims.Role))
		return c.Next()
	}
}

func RequireRole(roles ...auth.Role) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[string(role)] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("userRole").(string)
		if _, ok := allowed[role]; !ok {
			return response.Fail(c, apperror.ErrForbidden)
		}
		return c.Next()
	}
}
