package api

import (
	"context"
	"log/slog"
	"time"

	"temp_backend/config"
	"temp_backend/internal/handler"
	"temp_backend/internal/middleware"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Server wraps the Fiber application and its dependencies.
type Server struct {
	app    *fiber.App
	cfg    config.Config
	logger *slog.Logger
}

// NewServer builds a configured Fiber server and wires all routes.
func NewServer(cfg config.Config, logger *slog.Logger, itemHandler *handler.ItemHandler, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, roomHandler *handler.RoomHandler, postHandler *handler.PostHandler, commentHandler *handler.CommentHandler, authService service.AuthService) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: middleware.NewErrorHandler(logger),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    100 * 1024 * 1024, // 100 MiB
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(flogger.New(flogger.Config{
		Format: "[${time}] ${status} ${method} ${path} (${latency}) ${ip}\n",
	}))

	s := &Server{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
	s.registerRoutes(itemHandler, userHandler, authHandler, roomHandler, postHandler, commentHandler, authService)
	return s
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	return s.app.Listen(":" + s.cfg.App.Port)
}

// Shutdown gracefully stops the server using the supplied timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithTimeout(15 * time.Second)
}

// GetApp returns the underlying Fiber application for testing purposes.
func (s *Server) GetApp() *fiber.App {
	return s.app
}
