//go:build wireinject
// +build wireinject

package main

import (
	"main/config"
	"main/handlers"
	"main/repository"
	"main/router"
	"main/scheduler"
	"main/services"
	chat "main/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Definisikan struct untuk menampung hasil inisialisasi
type Application struct {
	App  *fiber.App
	Cron *cron.Cron
}

// Buat provider untuk struct Application
func NewApplication(app *fiber.App, cron *cron.Cron) *Application {
	return &Application{App: app, Cron: cron}
}

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
	services.NewUploadService,
)

var handlerSet = wire.NewSet(
	handlers.NewAuthHandler,
	handlers.NewChatHandler,
	handlers.NewUserHandler,
	handlers.NewRoomHandler,
	handlers.NewUploadHandler,
)

var schedulerSet = wire.NewSet(
	scheduler.InitCronJobs,
)

// InitializeApp sekarang mengembalikan struct Application dan error
func InitializeApp(cfg *config.Config, db *gorm.DB, hub *chat.Hub) (*Application, error) {
	wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		schedulerSet,
		router.NewRouter,
		NewApplication, // <-- Tambahkan provider NewApplication
	)
	return nil, nil
}
