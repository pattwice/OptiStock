package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"optistock/pkg/apperror"
)

type Envelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Meta    map[string]any `json:"meta,omitempty"`
	Error   *ErrorBody     `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Envelope{Success: true, Data: data, Error: nil})
}

func OKWithMeta(c *fiber.Ctx, data any, meta map[string]any) error {
	return c.JSON(Envelope{Success: true, Data: data, Meta: meta, Error: nil})
}

func Fail(c *fiber.Ctx, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.HTTPStatus).JSON(Envelope{
			Success: false,
			Error: &ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    apperror.ErrInternal.Code,
			Message: apperror.ErrInternal.Message,
		},
	})
}
