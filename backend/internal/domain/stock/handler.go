package stock

import (
	"github.com/gofiber/fiber/v2"
	"optistock/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/on-hand", h.ListOnHand)
}

func (h *Handler) ListOnHand(c *fiber.Ctx) error {
	itemCode := c.Query("item_code")
	status := c.Query("status")
	rows, err := h.service.ListOnHand(c.Context(), itemCode, status)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}
