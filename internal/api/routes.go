package api

import (
	"github.com/gofiber/fiber/v2"
	"temp_backend/internal/handler"
	"temp_backend/internal/middleware"
	"temp_backend/internal/service"
)

func (s *Server) registerRoutes(itemHandler *handler.ItemHandler, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, roomHandler *handler.RoomHandler, authService service.AuthService) {
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Public auth routes
	authGroup := s.app.Group("/api/v1/auth")
	authGroup.Post("/register", userHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.RefreshAccessToken)

	api := s.app.Group("/api/v1")

	// Protected routes with JWT middleware
	apiProtected := api.Group("", middleware.AuthMiddleware(authService))

	// User routes (protected)
	apiProtected.Get("/users/:id", userHandler.GetUser)
	apiProtected.Get("/profile", userHandler.ViewProfile)
	apiProtected.Patch("/profile", userHandler.UpdateProfile)
	apiProtected.Post("/profile/change-password", userHandler.ChangePassword)
	apiProtected.Delete("/profile", userHandler.DeleteAccount)
	apiProtected.Delete("/users/:id", userHandler.DeleteUser)
	apiProtected.Post("/auth/logout", authHandler.Logout)

	// Room routes (protected)
	apiProtected.Post("/rooms", roomHandler.CreateRoom)
	apiProtected.Post("/rooms/join", roomHandler.AddUserToRoom)
	apiProtected.Get("/rooms/:code", roomHandler.GetRoom)
	apiProtected.Get("/rooms/:code/members", roomHandler.GetRoomMembers)

	// Item routes (protected)
	apiProtected.Post("/items", itemHandler.CreateItem)
	apiProtected.Get("/items", itemHandler.ListItems)
	apiProtected.Get("/items/:id", itemHandler.GetItem)
	apiProtected.Put("/items/:id", itemHandler.UpdateItem)
	apiProtected.Delete("/items/:id", itemHandler.DeleteItem)
	apiProtected.Get("/items/:id/download", itemHandler.DownloadItem)
}
