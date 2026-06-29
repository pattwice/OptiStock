package workorder

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"optistock/internal/domain/alert"
	"optistock/pkg/apperror"
	"optistock/pkg/response"
)

type Handler struct {
	service *Service
	alerts  alert.Notifier
}

func NewHandler(service *Service, alerts alert.Notifier) *Handler {
	return &Handler{service: service, alerts: alerts}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
	router.Post("/", h.Create)
	router.Get("/:woNumber/allocation-proposal", h.AllocationProposal)
	router.Get("/:woNumber/audit", h.AuditLog)
	router.Post("/:woNumber/reserve", h.Reserve)
	router.Post("/:woNumber/start", h.StartProduction)
	router.Patch("/:woNumber/actuals", h.UpdateActuals)
	router.Post("/:woNumber/resolve-overusage", h.ResolveOverUsage)
	router.Post("/:woNumber/submit-for-approval", h.SubmitForApproval)
	router.Post("/:woNumber/complete", h.Complete)
	router.Post("/:woNumber/cancel", h.Cancel)
	router.Post("/:woNumber/reopen", h.Reopen)
	router.Get("/:woNumber", h.Get)
}

func (h *Handler) List(c *fiber.Ctx) error {
	rows, err := h.service.List(c.Context(), c.Query("status"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input CreateWOInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.Create(c.Context(), userID, input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	wo, err := h.service.Get(c.Context(), c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) AllocationProposal(c *fiber.Ctx) error {
	rows, err := h.service.AllocationProposal(c.Context(), c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}

func (h *Handler) Reserve(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input ReserveInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.Reserve(c.Context(), userID, c.Params("woNumber"), input)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Code == apperror.ErrInsufficientStock.Code && h.alerts != nil {
			h.alerts.OnReservationBlocked(c.Context(), c.Params("woNumber"), appErr.Details)
		}
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) StartProduction(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	wo, err := h.service.StartProduction(c.Context(), userID, c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) UpdateActuals(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input UpdateActualsInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.UpdateActuals(c.Context(), userID, c.Params("woNumber"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) ResolveOverUsage(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input ResolveOverUsageInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.ResolveOverUsage(c.Context(), userID, c.Params("woNumber"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) SubmitForApproval(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input CompleteInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.SubmitForApproval(c.Context(), userID, c.Params("woNumber"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	var input CompleteInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, err)
	}
	wo, err := h.service.Complete(c.Context(), userID, c.Params("woNumber"), input)
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	wo, err := h.service.Cancel(c.Context(), userID, c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) Reopen(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	wo, err := h.service.Reopen(c.Context(), userID, c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, wo)
}

func (h *Handler) AuditLog(c *fiber.Ctx) error {
	rows, err := h.service.AuditLog(c.Context(), c.Params("woNumber"))
	if err != nil {
		return response.Fail(c, err)
	}
	return response.OK(c, rows)
}
