package api

import (
	"temp_backend/internal/handler"
	"temp_backend/internal/middleware"
	"temp_backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) registerRoutes(itemHandler *handler.ItemHandler, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, roomHandler *handler.RoomHandler, postHandler *handler.PostHandler, commentHandler *handler.CommentHandler, authService service.AuthService) {
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// DEBUG: returns all environment variables visible to the process.
	// This endpoint is intended for local testing only and must not be deployed to production.
	/*
		s.app.Get("/debug/env", func(c *fiber.Ctx) error {
			envVars := make(map[string]string)
			for _, env := range os.Environ() {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					envVars[parts[0]] = parts[1]
				}
			}
			return c.JSON(fiber.Map{
				"env": envVars,
			})
		})
	*/

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
	apiProtected.Post("/rooms/add-member-by-code", roomHandler.AddUserToRoomByUserCode)
	apiProtected.Get("/rooms", roomHandler.ListUserRooms)
	apiProtected.Get("/rooms/by-id/:id", roomHandler.GetRoomByID)
	apiProtected.Get("/rooms/:code", roomHandler.GetRoom)
	apiProtected.Get("/rooms/:code/members", roomHandler.GetRoomMembers)
	apiProtected.Get("/rooms/:code/users", roomHandler.GetRoomUsers)
	apiProtected.Get("/rooms/:code/posts", postHandler.GetPostsByRoomCode)
	apiProtected.Post("/rooms/:code/remove-member", roomHandler.RemoveMemberFromRoom)
	apiProtected.Post("/rooms/:code/leave", roomHandler.LeaveRoom)
	apiProtected.Delete("/rooms/:code", roomHandler.HandleRoomDelete)

	// Post routes (protected)
	apiProtected.Post("/posts", postHandler.CreatePost)
	apiProtected.Get("/posts/:id", postHandler.GetPost)
	apiProtected.Delete("/posts/:id", postHandler.DeletePost)

	// Comment routes (protected)
	apiProtected.Post("/posts/:id/comments", commentHandler.CreateComment)
	apiProtected.Get("/posts/:id/comments", commentHandler.GetCommentsByPostID)
	apiProtected.Delete("/comments/:id", commentHandler.DeleteComment)

	// Item routes (protected)
	apiProtected.Post("/items", itemHandler.CreateItem)
	apiProtected.Get("/items", itemHandler.ListItems)
	apiProtected.Get("/items/:id", itemHandler.GetItem)
	apiProtected.Put("/items/:id", itemHandler.UpdateItem)
	apiProtected.Delete("/items/:id", itemHandler.DeleteItem)
	apiProtected.Get("/items/:id/download", itemHandler.DownloadItem)
}
