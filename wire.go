//go:build wireinject
// +build wireinject

package main

import (
	"main/config"
	"main/handlers"
	"main/repository"
	"main/router"
	"main/services"
	chat "main/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var repositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewTokenRepository,
	repository.NewRoomRepository,
	repository.NewChatRepository,
)

var serviceSet = wire.NewSet(
	services.NewJWTService,
	services.NewAuthService,
	services.NewUserService,
	services.NewRoomService,
)

var handlerSet = wire.NewSet(
	handlers.NewAuthHandler,
	handlers.NewChatHandler,
	handlers.NewUserHandler,
	handlers.NewRoomHandler,
)


func InitializeApp(cfg *config.Config, db *gorm.DB, hub *chat.Hub) *fiber.App {
	wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		router.NewRouter,
	)
	return nil
}