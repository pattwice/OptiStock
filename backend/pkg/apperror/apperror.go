package apperror

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

var (
	ErrUnauthorized          = New("UNAUTHORIZED", "authentication required", http.StatusUnauthorized)
	ErrForbidden             = New("FORBIDDEN", "insufficient permissions", http.StatusForbidden)
	ErrInvalidCredentials    = New("INVALID_CREDENTIALS", "invalid email or password", http.StatusUnauthorized)
	ErrInvalidToken          = New("INVALID_TOKEN", "invalid or expired token", http.StatusUnauthorized)
	ErrNotFound              = New("NOT_FOUND", "resource not found", http.StatusNotFound)
	ErrValidation            = New("VALIDATION_ERROR", "validation failed", http.StatusUnprocessableEntity)
	ErrInternal              = New("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
	ErrInsufficientStock     = New("INSUFFICIENT_STOCK", "available stock cannot satisfy reservation", http.StatusConflict)
	ErrNegativeStockPrevented = New("NEGATIVE_STOCK_PREVENTED", "transaction would drop physical stock below zero", http.StatusConflict)
	ErrLotNotEligible        = New("LOT_NOT_ELIGIBLE", "lot is hold or quarantined", http.StatusConflict)
	ErrWOStatusInvalid       = New("WO_STATUS_INVALID", "requested status transition is not allowed", http.StatusUnprocessableEntity)
	ErrApprovalRequired      = New("APPROVAL_REQUIRED", "supervisor approval required", http.StatusForbidden)
	ErrCompletionGuardFailed = New("COMPLETION_GUARD_FAILED", "actual plus damage exceeds reserved quantity", http.StatusUnprocessableEntity)
)

func WithDetails(err *AppError, details any) *AppError {
	copy := *err
	copy.Details = details
	return &copy
}

func WithMessage(err *AppError, message string) *AppError {
	copy := *err
	copy.Message = message
	return &copy
}
