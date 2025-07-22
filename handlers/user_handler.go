package handlers

import (
	"errors"
	"main/dto"
	"main/services"
	"main/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}

	userResponse, err := h.userService.CreateUser(req)
	if err != nil {
		if err.Error() == "email sudah terdaftar" {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal membuat pengguna")
	}

	return c.Status(fiber.StatusCreated).JSON(utils.BaseResponse{
		Success: true,
		Message: "Pengguna berhasil dibuat",
		Data:    userResponse,
	})
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data pengguna")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Data semua pengguna berhasil diambil",
		Data:    users,
	})
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Pengguna tidak ditemukan")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data pengguna")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Data pengguna berhasil diambil",
		Data:    user,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}

	updatedUser, err := h.userService.UpdateUser(uint(id), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Pengguna tidak ditemukan untuk diperbarui")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal memperbarui pengguna")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Pengguna berhasil diperbarui",
		Data:    updatedUser,
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Pengguna tidak ditemukan untuk dihapus")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menghapus pengguna")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Pengguna berhasil dihapus",
	})
}

func (h *UserHandler) GetMyProfile(c *fiber.Ctx) error {
	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token")
	}

	userID, ok := userIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pengguna tidak valid di context")
	}

	userResponse, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Pengguna tidak ditemukan")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data profil")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Profil berhasil diambil",
		Data:    userResponse,
	})
}

func (h *UserHandler) UpdateMyProfile(c *fiber.Ctx) error {
	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token")
	}
	userID, ok := userIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pengguna tidak valid di context")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}

	updatedUser, err := h.userService.UpdateUser(uint(userID), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Pengguna tidak ditemukan untuk diperbarui")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal memperbarui profil")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Profil berhasil diperbarui",
		Data:    updatedUser,
	})
}