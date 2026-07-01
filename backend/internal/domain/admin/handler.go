package admin

import (
	"github.com/gofiber/fiber/v2"
	"optistock/pkg/apperror"
	"optistock/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/users", h.ListUsers)
	router.Post("/users", h.CreateUser)
	router.Patch("/users/:userID", h.UpdateUser)
	router.Get("/config", h.ListConfig)
	router.Patch("/config", h.PatchConfig)
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.service.ListUsers(c.Context())
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, users)
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var input CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, apperror.WithMessage(apperror.ErrValidation, "invalid JSON body"))
	}
	user, err := h.service.CreateUser(c.Context(), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, user)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var input UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, apperror.WithMessage(apperror.ErrValidation, "invalid JSON body"))
	}
	user, err := h.service.UpdateUser(c.Context(), c.Params("userID"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, user)
}

func (h *Handler) ListConfig(c *fiber.Ctx) error {
	entries, err := h.service.ListConfig(c.Context())
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, entries)
}

func (h *Handler) PatchConfig(c *fiber.Ctx) error {
	var updates map[string]string
	if err := c.BodyParser(&updates); err != nil {
		return response.Fail(c, apperror.WithMessage(apperror.ErrValidation, "invalid JSON body"))
	}
	entries, err := h.service.PatchConfig(c.Context(), updates)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, entries)
}
