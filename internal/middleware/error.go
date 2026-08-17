package middleware

import (
	"errors"
	"log/slog"

	"temp_backend/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// NewErrorHandler returns a centralized Fiber error handler that maps errors
// to structured JSON responses using AppError. When showInternal is true
// (e.g. in development), wrapped internal errors are included in the response
// as a "detail" field for debugging.
func NewErrorHandler(log *slog.Logger, showInternal bool) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Extract AppError from the error chain.
		var appErr *domain.AppError
		if !errors.As(err, &appErr) {
			// Allow Fiber's own *fiber.Error to set the status code and message.
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"error":  fiberErr.Message,
					"code":   "HTTP_ERROR",
					"status": fiberErr.Code,
				})
			}
			appErr = domain.ErrInternalServer
		}

		resp := fiber.Map{
			"error":  appErr.Message,
			"code":   appErr.Code,
			"status": appErr.HTTPStatus,
		}

		// In development, include the wrapped internal error for debugging.
		if showInternal && appErr.Err != nil {
			resp["detail"] = appErr.Err.Error()
		}

		if appErr.HTTPStatus >= 500 {
			log.Error("request failed",
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", appErr.HTTPStatus),
				slog.String("code", appErr.Code),
				slog.String("error", appErr.Error()),
			)
		}

		return c.Status(appErr.HTTPStatus).JSON(resp)
	}
}