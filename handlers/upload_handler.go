package handlers

import (
	"main/services"
	"main/utils"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	uploadService services.UploadService
}

func NewUploadHandler(uploadService services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadImage adalah handler untuk endpoint POST /api/upload
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	// 1. Ambil file dari form request. "image" adalah nama field di form.
	file, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "File gambar tidak ditemukan di request")
	}

	// 2. Panggil service untuk mengunggah file.
	url, err := h.uploadService.UploadFile(file, "images", file.Filename)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengunggah file")
	}
	
	// 3. Kirim kembali URL gambar sebagai respons.
	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "File berhasil diunggah",
		Data:    fiber.Map{"url": url},
	})
}