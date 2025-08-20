package handlers

import (
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
