package api

import (
	"github.com/gofiber/fiber/v2"
	"temp_backend/internal/handler"
)

func (s *Server) registerRoutes(itemHandler *handler.ItemHandler) {
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	api := s.app.Group("/api/v1")

	api.Post("/items", itemHandler.CreateItem)
	api.Get("/items", itemHandler.ListItems)
	api.Get("/items/:id", itemHandler.GetItem)
	api.Put("/items/:id", itemHandler.UpdateItem)
	api.Delete("/items/:id", itemHandler.DeleteItem)
	api.Get("/items/:id/download", itemHandler.DownloadItem)
}
