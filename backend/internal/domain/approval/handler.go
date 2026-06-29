package approval

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
	router.Get("/", h.ListPending)
	router.Post("/:approvalID/withdraw", h.Withdraw)
	router.Post("/:approvalID/approve", h.Approve)
	router.Post("/:approvalID/reject", h.Reject)
}

func (h *Handler) ListPending(c *fiber.Ctx) error {
	rows, err := h.service.ListPending(c.Context())
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) ListByWO(c *fiber.Ctx) error {
	rows, err := h.service.ListByWO(c.Context(), c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) Withdraw(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	row, err := h.service.Withdraw(c.Context(), userID, c.Params("approvalID"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, row)
}

func (h *Handler) Approve(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input ResolveInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	row, err := h.service.Approve(c.Context(), userID, c.Params("approvalID"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, row)
}

func (h *Handler) Reject(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input ResolveInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	row, err := h.service.Reject(c.Context(), userID, c.Params("approvalID"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, row)
}
