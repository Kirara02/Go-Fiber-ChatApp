package handlers

import (
	"fmt"
	"io"
	"log"
	"main/dto"
	"main/services"
	"main/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AIHandler struct {
	service services.AIService
}

func NewAIHandler(service services.AIService) *AIHandler {
	return &AIHandler{service: service}
}

func (h *AIHandler) ChatWithAI(c *fiber.Ctx) error {
	var req dto.AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Request body tidak valid")
	}

	if req.Message == "" {
		return fiber.NewError(http.StatusBadRequest, "Pesan tidak boleh kosong")
	}

	ctx := c.Context()

	aiResponse, err := h.service.GenerateContent(ctx, req.Message)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Gagal mendapatkan respons dari AI: "+err.Error())
	}

	return c.Status(http.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Respons AI berhasil didapatkan",
		Data:    aiResponse,
	})
}

func (h *AIHandler) ChatStream(c *fiber.Ctx) error {
	var req dto.AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Request body tidak valid")
	}

	if req.Message == "" {
		return fiber.NewError(http.StatusBadRequest, "Pesan tidak boleh kosong")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	streamChan, err := h.service.GenerateContentStream(c.Context(), req.Message)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Gagal memulai stream AI: "+err.Error())
	}

	// 1. Buat io.Pipe
	// pipeReader akan diberikan ke Fiber, pipeWriter akan kita gunakan untuk menulis data
	pipeReader, pipeWriter := io.Pipe()

	// 2. Jalankan goroutine untuk menangani penulisan ke pipeWriter
	go func() {
		// Pastikan pipeWriter ditutup saat goroutine selesai, ini akan memberi sinyal EOF ke pipeReader
		defer pipeWriter.Close()

		// Baca data dari channel AI service
		for chunk := range streamChan {
			// Format sebagai Server-Sent Event (SSE)
			sseMessage := fmt.Sprintf("data: %s\n\n", chunk)

			// Tulis pesan SSE ke pipeWriter
			_, err := fmt.Fprint(pipeWriter, sseMessage)
			if err != nil {
				log.Printf("Error writing to pipe: %v", err)
				return // Hentikan goroutine jika ada error (misal: client menutup koneksi)
			}
		}
	}()

	// 3. Berikan pipeReader ke SendStream
	// SendStream akan membaca dari pipeReader sampai mendapatkan EOF (saat pipeWriter ditutup)
	return c.SendStream(pipeReader)
}
