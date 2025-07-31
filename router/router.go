package router

import (
	"main/config"
	"main/handlers"
	"main/middleware"
	"main/repository"

	"main/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	ws "github.com/gofiber/websocket/v2"
)

func NewRouter(
	chatHandler *handlers.ChatHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	roomHandler *handlers.RoomHandler,
	uploadHandler *handlers.UploadHandler,
	tokenRepo repository.TokenRepository,
	cfg *config.Config,
) *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Terjadi kesalahan internal pada server"

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(utils.BaseResponse{
				Success: false,
				Message: message,
				Error: &utils.ErrorResponse{
					Code:    code,
					Details: err.Error(),
				},
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "*",
	}))
	app.Static("/", "./web")

	api := app.Group("/api")

	// "/auth" Route
	authRoute := api.Group("/auth")
	authRoute.Post("/register", authHandler.Register)
	authRoute.Post("/login", authHandler.Login)
	authRoute.Post("/refresh", authHandler.RefreshToken)
	authRoute.Post("/logout", middleware.AuthMiddleware(cfg, tokenRepo), authHandler.Logout)
	authRoute.Put("/change-password", middleware.AuthMiddleware(cfg, tokenRepo), authHandler.ChangePassword)

	// Route with authentication
	auth := middleware.AuthMiddleware(cfg, tokenRepo)
	api.Use(auth)

	api.Get("/profile", userHandler.GetMyProfile)
	api.Put("/profile", userHandler.UpdateMyProfile)

	api.Post("/upload", uploadHandler.UploadImage)

	userRoutes := api.Group("/users")
	userRoutes.Post("/", userHandler.CreateUser)
	userRoutes.Get("/", userHandler.GetAllUsers)
	userRoutes.Get("/:id", userHandler.GetUserByID)
	userRoutes.Put("/:id", userHandler.UpdateUser)
	userRoutes.Delete("/:id", userHandler.DeleteUser)

	roomRoutes := api.Group("/rooms")
	roomRoutes.Post("/", roomHandler.CreateRoom)
	roomRoutes.Get("/", roomHandler.GetMyRooms)
	roomRoutes.Get("/:id", roomHandler.GetRoomByID)
	roomRoutes.Put("/:id/image", roomHandler.UpdateRoomImage)

	// WebSocket
	chatRoutes := app.Group("/chat")
	chatRoutes.Use(auth)
	chatRoutes.Get("/ws/:roomId", ws.New(chatHandler.HandleWebSocket))

	return app
}
