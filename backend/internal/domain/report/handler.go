package report

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
	router.Get("/stock-on-hand", h.StockOnHand)
	router.Get("/stock-on-hand/export", h.ExportStockOnHand)
	router.Get("/movement-ledger", h.MovementLedger)
	router.Get("/movement-ledger/export", h.ExportMovementLedger)
	router.Get("/wo-summary", h.WOSummary)
	router.Get("/wo-summary/export", h.ExportWOSummary)
	router.Get("/shortage-damage", h.ShortageDamage)
	router.Get("/shortage-damage/export", h.ExportShortageDamage)
	router.Get("/audit-trail", h.AuditTrail)
	router.Get("/audit-trail/export", h.ExportAuditTrail)
	router.Get("/partial-completion", h.PartialCompletion)
	router.Get("/partial-completion/export", h.ExportPartialCompletion)
}

func (h *Handler) StockOnHand(c *fiber.Ctx) error {
	result, err := h.service.StockOnHand(c.Context(), StockOnHandFilters{
		ItemType:   c.Query("item_type"),
		LotStatus:  c.Query("lot_status"),
		NearExpiry: c.Query("near_expiry") == "true",
		ItemCode:   c.Query("item_code"),
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportStockOnHand(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportStockOnHand(c.Context(), StockOnHandFilters{
			ItemType:   c.Query("item_type"),
			LotStatus:  c.Query("lot_status"),
			NearExpiry: c.Query("near_expiry") == "true",
			ItemCode:   c.Query("item_code"),
		}, c.Query("format"))
	})
}

func (h *Handler) MovementLedger(c *fiber.Ctx) error {
	result, err := h.service.MovementLedger(c.Context(), MovementLedgerFilters{
		FromDate:        c.Query("from_date"),
		ToDate:          c.Query("to_date"),
		TransactionType: c.Query("transaction_type"),
		ItemCode:        c.Query("item_code"),
		LotInternalID:   c.Query("lot_internal_id"),
		Page:            queryInt(c, "page", 1),
		PageSize:        queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportMovementLedger(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportMovementLedger(c.Context(), MovementLedgerFilters{
			FromDate:        c.Query("from_date"),
			ToDate:          c.Query("to_date"),
			TransactionType: c.Query("transaction_type"),
			ItemCode:        c.Query("item_code"),
			LotInternalID:   c.Query("lot_internal_id"),
		}, c.Query("format"))
	})
}

func (h *Handler) WOSummary(c *fiber.Ctx) error {
	result, err := h.service.WOSummary(c.Context(), WOSummaryFilters{
		Status:       c.Query("status"),
		FromDate:     c.Query("from_date"),
		ToDate:       c.Query("to_date"),
		TargetFGCode: c.Query("target_fg_code"),
		Page:         queryInt(c, "page", 1),
		PageSize:     queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportWOSummary(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportWOSummary(c.Context(), WOSummaryFilters{
			Status:       c.Query("status"),
			FromDate:     c.Query("from_date"),
			ToDate:       c.Query("to_date"),
			TargetFGCode: c.Query("target_fg_code"),
		}, c.Query("format"))
	})
}

func (h *Handler) ShortageDamage(c *fiber.Ctx) error {
	result, err := h.service.ShortageDamage(c.Context(), ShortageDamageFilters{
		FromDate:   c.Query("from_date"),
		ToDate:     c.Query("to_date"),
		ReasonCode: c.Query("reason_code"),
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportShortageDamage(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportShortageDamage(c.Context(), ShortageDamageFilters{
			FromDate:   c.Query("from_date"),
			ToDate:     c.Query("to_date"),
			ReasonCode: c.Query("reason_code"),
		}, c.Query("format"))
	})
}

func (h *Handler) AuditTrail(c *fiber.Ctx) error {
	result, err := h.service.AuditTrail(c.Context(), AuditTrailFilters{
		WONumber: c.Query("wo_number"),
		Action:   c.Query("action"),
		UserID:   c.Query("user_id"),
		FromDate: c.Query("from_date"),
		ToDate:   c.Query("to_date"),
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportAuditTrail(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportAuditTrail(c.Context(), AuditTrailFilters{
			WONumber: c.Query("wo_number"),
			Action:   c.Query("action"),
			UserID:   c.Query("user_id"),
			FromDate: c.Query("from_date"),
			ToDate:   c.Query("to_date"),
		}, c.Query("format"))
	})
}

func (h *Handler) PartialCompletion(c *fiber.Ctx) error {
	result, err := h.service.PartialCompletion(c.Context(), PartialCompletionFilters{
		FromDate:       c.Query("from_date"),
		ToDate:         c.Query("to_date"),
		ApprovalStatus: c.Query("approval_status"),
		Page:           queryInt(c, "page", 1),
		PageSize:       queryInt(c, "page_size", 50),
	})
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, result)
}

func (h *Handler) ExportPartialCompletion(c *fiber.Ctx) error {
	return h.sendExport(c, func() ([]byte, string, string, error) {
		return h.service.ExportPartialCompletion(c.Context(), PartialCompletionFilters{
			FromDate:       c.Query("from_date"),
			ToDate:         c.Query("to_date"),
			ApprovalStatus: c.Query("approval_status"),
		}, c.Query("format"))
	})
}

func (h *Handler) sendExport(c *fiber.Ctx, fn func() ([]byte, string, string, error)) error {
	data, filename, contentType, err := fn()
	if err != nil {
		return response.Fail(c, err)
	}
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Send(data)
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	v := c.Query(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
