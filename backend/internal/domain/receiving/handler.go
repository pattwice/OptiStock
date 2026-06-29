package receiving

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
	router.Post("/po", h.POReceipt)
	router.Post("/adjustment", h.Adjustment)
}

func (h *Handler) POReceipt(c *fiber.Ctx) error {
	var input POReceiptInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	userID, _ := c.Locals("userID").(string)
	out, err := h.service.POReceipt(c.Context(), userID, input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, out)
}

func (h *Handler) Adjustment(c *fiber.Ctx) error {
	var input AdjustmentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	userID, _ := c.Locals("userID").(string)
	if err := h.service.Adjustment(c.Context(), userID, input); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"status": "ok"})
}

