package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"main/dto"
	"main/services"
	"main/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AIHandler sekarang bergantung pada SessionService dan AIService.
type AIHandler struct {
	aiService      services.AIService      // Dependensi untuk logika AI
	sessionService services.SessionService // Dependensi untuk logika Sesi
}

// NewAIHandler diperbarui untuk menerima SessionService.
func NewAIHandler(aiService services.AIService, sessionService services.SessionService) *AIHandler {
	return &AIHandler{
		aiService:      aiService,
		sessionService: sessionService,
	}
}

// ChatWithAI (Non-Streaming) sekarang menggunakan service untuk mengelola state.
func (h *AIHandler) ChatWithAI(c *fiber.Ctx) error {
	var req dto.AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Request body tidak valid")
	}
	if req.SessionID == "" || req.Message == "" {
		return fiber.NewError(http.StatusBadRequest, "sessionId dan message tidak boleh kosong")
	}

	// --- PERBAIKAN DI SINI ---
	// Mengambil userID dari c.Locals dengan cara yang lebih aman.
	var userID uint
	userIDValue := c.Locals("user_id")

	if userIDValue != nil {
		// JWT claims seringkali mendekode angka sebagai float64, jadi kita tangani kasus itu.
		if idFloat, ok := userIDValue.(float64); ok {
			userID = uint(idFloat)
		} else if idUint, ok := userIDValue.(uint); ok {
			// Menangani kasus jika tipenya sudah benar uint.
			userID = idUint
		} else {
			// Log jika tipe datanya tidak terduga.
			log.Printf("PERINGATAN: Gagal melakukan type assertion untuk user_id. Tipe data aktual: %T", userIDValue)
		}
	} else {
		log.Printf("PERINGATAN: c.Locals(\"user_id\") tidak ditemukan.")
	}

	// Log untuk debugging
	log.Printf("DEBUG: UserID yang akan digunakan untuk sesi: %d", userID)

	// 1. Dapatkan atau buat sesi dari service.
	session, err := h.sessionService.GetOrCreateSession(req.SessionID, userID)
	if err != nil {
		log.Printf("ERROR: Gagal mendapatkan sesi dari service: %v", err)
		return fiber.NewError(http.StatusInternalServerError, "Gagal memuat state percakapan")
	}

	log.Printf("INFO: [SessionID: %s] Menggunakan ringkasan: \"%s\"", req.SessionID, session.Summary)

	// 2. Panggil service AI dengan ringkasan dari sesi.
	aiResponse, err := h.aiService.GenerateContent(c.Context(), session.Summary, req.Message)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Gagal mendapatkan respons dari AI: "+err.Error())
	}

	// 3. Perbarui ringkasan melalui service.
	err = h.sessionService.UpdateSummary(session, aiResponse.NewSummary)
	if err != nil {
		log.Printf("ERROR: Gagal memperbarui ringkasan melalui service: %v", err)
	}

	log.Printf("INFO: [SessionID: %s] Ringkasan diperbarui menjadi: \"%s\"", session.SessionID, aiResponse.NewSummary)

	// Hapus NewSummary dari respons agar tidak dikirim ke klien.
	aiResponse.NewSummary = ""

	return c.Status(http.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Respons AI berhasil didapatkan",
		Data:    aiResponse,
	})
}

// ChatStream sekarang diimplementasikan dengan memori dan summarization.
func (h *AIHandler) ChatStream(c *fiber.Ctx) error {
	var req dto.AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(http.StatusBadRequest, "Request body tidak valid")
	}
	if req.SessionID == "" || req.Message == "" {
		return fiber.NewError(http.StatusBadRequest, "sessionId dan message tidak boleh kosong")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	userID, _ := c.Locals("user_id").(uint)

	// 1. Dapatkan sesi dari service.
	session, err := h.sessionService.GetOrCreateSession(req.SessionID, userID)
	if err != nil {
		log.Printf("ERROR: Gagal mendapatkan sesi dari service: %v", err)
		return fiber.NewError(http.StatusInternalServerError, "Gagal memuat state percakapan")
	}

	log.Printf("INFO: [SessionID: %s] Memulai stream dengan ringkasan: \"%s\"", req.SessionID, session.Summary)

	// 2. Panggil service streaming.
	streamChan, fullResponseChan, err := h.aiService.GenerateContentStream(c.Context(), session.Summary, req.Message)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, "Gagal memulai stream AI: "+err.Error())
	}

	// 3. Jalankan goroutine di latar belakang untuk memperbarui ringkasan setelah stream selesai.
	go func() {
		// Tunggu respons lengkap dari service.
		fullResponse := <-fullResponseChan

		if fullResponse == "" {
			log.Printf("PERINGATAN: [SessionID: %s] Respons penuh dari stream kosong, ringkasan tidak diperbarui.", req.SessionID)
			return
		}

		// Panggil service untuk membuat ringkasan baru.
		newSummary, err := h.aiService.Summarize(context.Background(), session.Summary, req.Message, fullResponse)
		if err != nil {
			log.Printf("ERROR: [SessionID: %s] Gagal meringkas stream: %v", req.SessionID, err)
			return
		}

		// Perbarui ringkasan di database.
		err = h.sessionService.UpdateSummary(session, newSummary)
		if err != nil {
			log.Printf("ERROR: [SessionID: %s] Gagal memperbarui ringkasan stream di DB: %v", req.SessionID, err)
		}

		log.Printf("INFO: [SessionID: %s] Ringkasan dari stream diperbarui menjadi: \"%s\"", req.SessionID, newSummary)
	}()

	// 4. Salurkan respons stream ke klien.
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()
		for chunk := range streamChan {
			sseMessage := fmt.Sprintf("data: %s\n\n", chunk)
			if _, err := fmt.Fprint(pipeWriter, sseMessage); err != nil {
				log.Printf("Error writing to pipe: %v", err)
				return
			}
		}
	}()

	return c.SendStream(pipeReader)
}
