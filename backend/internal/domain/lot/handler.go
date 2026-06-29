package lot

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
	router.Get("/", h.ListLots)
	router.Post("/", h.CreateLot)
	router.Get("/:lotInternalID", h.GetLot)
	router.Patch("/:lotInternalID/status", h.UpdateLotStatus)
}

func (h *Handler) ListLots(c *fiber.Ctx) error {
	itemCode := c.Query("item_code")
	status := c.Query("status")
	rows, err := h.service.ListLots(c.Context(), itemCode, status)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) GetLot(c *fiber.Ctx) error {
	lotID := c.Params("lotInternalID")
	l, err := h.service.GetLot(c.Context(), lotID)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, l)
}

func (h *Handler) CreateLot(c *fiber.Ctx) error {
	var input CreateLotInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	l, err := h.service.CreateLot(c.Context(), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, l)
}

func (h *Handler) UpdateLotStatus(c *fiber.Ctx) error {
	lotID := c.Params("lotInternalID")
	var input UpdateLotStatusInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	l, err := h.service.UpdateLotStatus(c.Context(), lotID, input.Status)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, l)
}

