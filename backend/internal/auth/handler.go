package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"optistock/pkg/apperror"
	"optistock/pkg/response"
)

type Handler struct {
	service *Service
	secure  bool
}

func NewHandler(service *Service, appEnv string) *Handler {
	return &Handler{service: service, secure: appEnv != "development"}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/login", h.Login)
	router.Post("/refresh", h.Refresh)
	router.Post("/logout", h.Logout)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, apperror.WithMessage(apperror.ErrValidation, "invalid JSON body"))
	}
	if input.Email == "" || input.Password == "" {
		return response.Fail(c, apperror.WithMessage(apperror.ErrValidation, "email and password are required"))
	}

	result, err := h.service.Login(c.Context(), input)
	if err != nil {
		return response.Fail(c, err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	return response.OK(c, fiber.Map{
		"user":         result.User,
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(RefreshCookieName())
	result, err := h.service.Refresh(c.Context(), refreshToken)
	if err != nil {
		return response.Fail(c, err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	return response.OK(c, fiber.Map{
		"user":         result.User,
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies(RefreshCookieName())
	if err := h.service.Logout(c.Context(), refreshToken); err != nil {
		return response.Fail(c, err)
	}
	h.clearRefreshCookie(c)
	return response.OK(c, fiber.Map{"message": "logged out"})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return response.Fail(c, fiber.ErrUnauthorized)
	}
	user, err := h.service.Me(c.Context(), userID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, user)
}

func (h *Handler) setRefreshCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookieName(),
		Value:    token,
		HTTPOnly: true,
		Secure:   h.secure,
		SameSite: "Lax",
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *Handler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookieName(),
		Value:    "",
		HTTPOnly: true,
		Secure:   h.secure,
		SameSite: "Lax",
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}
