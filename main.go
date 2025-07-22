package main

import (
	"log"
	"main/config"
	"main/domain"
	chat "main/websocket"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Gagal memuat file .env")
	}

	cfg := config.New()
	
	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	log.Println("Koneksi database berhasil.")

	err = db.AutoMigrate(&domain.User{}, &domain.Room{}, &domain.ChatMessage{}, &domain.InvalidatedToken{})
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}
	log.Println("Migrasi database selesai.")

	hub := chat.NewHub()

	app := InitializeApp(cfg, db, hub)

	port := ":8080"
	log.Printf("Server Fiber memulai di port %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Gagal memulai server: %v", err)
	}
}