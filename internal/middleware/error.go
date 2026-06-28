package middleware

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"temp_backend/internal/domain"
)

// NewErrorHandler returns a centralized Fiber error handler that maps domain
// errors to HTTP status codes and emits structured logs.
func NewErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "internal server error"

		switch {
		case errors.Is(err, domain.ErrNotFound):
			code = fiber.StatusNotFound
			message = err.Error()
		case errors.Is(err, domain.ErrInvalidInput):
			code = fiber.StatusBadRequest
			message = err.Error()
		case errors.Is(err, domain.ErrConflict):
			code = fiber.StatusConflict
			message = err.Error()
		case errors.Is(err, domain.ErrUnauthorized):
			code = fiber.StatusUnauthorized
			message = err.Error()
		case errors.Is(err, domain.ErrForbidden):
			code = fiber.StatusForbidden
			message = err.Error()
		}

		// Allow Fiber's own *fiber.Error to set the status code.
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			message = fiberErr.Message
		}

		if code >= fiber.StatusInternalServerError {
			log.Error("request failed",
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", code),
				slog.String("error", err.Error()),
			)
		}

		return c.Status(code).JSON(fiber.Map{
			"error": message,
		})
	}
}
