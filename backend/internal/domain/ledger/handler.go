package ledger

import (
	"strconv"

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
	router.Get("/", h.List)
}

func (h *Handler) List(c *fiber.Ctx) error {
	itemCode := c.Query("item_code")
	lotID := c.Query("lot_internal_id")
	from := c.Query("from") // RFC3339
	to := c.Query("to")     // RFC3339

	limit := 0
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return response.Fail(c, err)
		}
		limit = n
	}

	rows, err := h.service.List(c.Context(), itemCode, lotID, from, to, limit)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

