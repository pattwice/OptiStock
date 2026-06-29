package item

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

func (h *Handler) RegisterItemRoutes(router fiber.Router) {
	router.Get("/", h.ListItems)
	router.Post("/", h.CreateItem)
	router.Get("/:itemCode", h.GetItem)
	router.Patch("/:itemCode", h.UpdateItem)
}

func (h *Handler) RegisterBOMRoutes(router fiber.Router) {
	router.Get("/", h.ListBOM)
	router.Post("/", h.CreateBOMRow)
	router.Post("/activate", h.ActivateBOMVersion)
}

func (h *Handler) ListItems(c *fiber.Ctx) error {
	itemType := c.Query("type")
	items, err := h.service.ListItems(c.Context(), itemType)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, items)
}

func (h *Handler) GetItem(c *fiber.Ctx) error {
	itemCode := c.Params("itemCode")
	it, err := h.service.GetItem(c.Context(), itemCode)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, it)
}

func (h *Handler) CreateItem(c *fiber.Ctx) error {
	var input CreateItemInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	it, err := h.service.CreateItem(c.Context(), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, it)
}

func (h *Handler) UpdateItem(c *fiber.Ctx) error {
	var input UpdateItemInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	itemCode := c.Params("itemCode")
	it, err := h.service.UpdateItem(c.Context(), itemCode, input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, it)
}

func (h *Handler) ListBOM(c *fiber.Ctx) error {
	parent := c.Query("parent_item_code")
	rows, err := h.service.ListBOMByParent(c.Context(), parent)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) CreateBOMRow(c *fiber.Ctx) error {
	var input CreateBOMRowInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	row, err := h.service.CreateBOMRow(c.Context(), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, row)
}

type activateBOMInput struct {
	ParentItemCode string `json:"parent_item_code"`
	BomVersion     string `json:"bom_version"`
}

func (h *Handler) ActivateBOMVersion(c *fiber.Ctx) error {
	var input activateBOMInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	if err := h.service.ActivateBOMVersion(c.Context(), input.ParentItemCode, input.BomVersion); err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, fiber.Map{"status": "ok"})
}

