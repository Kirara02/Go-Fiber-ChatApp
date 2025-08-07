package handlers

import (
	"errors" // Pastikan package 'errors' diimpor
	"fmt"
	"main/dto"
	"main/services"
	"main/utils"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Fungsi ini sudah bagus, tidak perlu diubah.
func formatValidationErrors(err error) map[string]string {
	errorMessages := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			fieldName := e.Field()
			switch e.Tag() {
			case "required":
				errorMessages[fieldName] = fmt.Sprintf("Field %s wajib diisi", fieldName)
			case "email":
				errorMessages[fieldName] = fmt.Sprintf("Field %s harus berupa format email yang valid", fieldName)
			case "min":
				errorMessages[fieldName] = fmt.Sprintf("Field %s harus memiliki minimal %s karakter", fieldName, e.Param())
			default:
				errorMessages[fieldName] = fmt.Sprintf("Field %s tidak valid", fieldName)
			}
		}
	}
	return errorMessages
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(dto.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Cannot parse JSON", Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: err.Error()},
		})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Data yang diberikan tidak valid",
			Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: formatValidationErrors(err)},
		})
	}

	user, err := h.authService.Register(req)
	if err != nil {
		// <-- Gunakan errors.Is untuk memeriksa tipe error
		if errors.Is(err, services.ErrEmailConflict) {
			return c.Status(fiber.StatusConflict).JSON(utils.BaseResponse{
				Success: false, Message: "Email ini sudah terdaftar",
				Error: &utils.ErrorResponse{Code: fiber.StatusConflict, Details: err.Error()},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
			Success: false, Message: "Gagal membuat pengguna", Error: &utils.ErrorResponse{Code: fiber.StatusInternalServerError, Details: err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(utils.BaseResponse{
		Success: true, Message: "Pengguna berhasil terdaftar", Data: dto.ToUserResponse(user),
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(dto.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Cannot parse JSON", Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: err.Error()},
		})
	}

	accessToken, refreshToken, user, err := h.authService.Login(req)
	// <-- Periksa error kredensial secara spesifik
	if errors.Is(err, services.ErrInvalidCredentials) {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Kredensial tidak valid",
			Error: &utils.ErrorResponse{Code: fiber.StatusUnauthorized, Details: "Email atau password salah"},
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
			Success: false, Message: "Terjadi kesalahan saat login", Error: &utils.ErrorResponse{Code: fiber.StatusInternalServerError, Details: err.Error()},
		})
	}

	loginData := dto.LoginResponse{
		User:         dto.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(utils.BaseResponse{
		Success: true, Message: "Login berhasil", Data: loginData,
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// (Fungsi ini tidak perlu diubah karena sudah menangani error secara umum)
	req := new(dto.RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Cannot parse JSON", Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: err.Error()},
		})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Refresh token wajib diisi",
			Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: formatValidationErrors(err)},
		})
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Gagal memperbarui token",
			Error: &utils.ErrorResponse{Code: fiber.StatusUnauthorized, Details: err.Error()},
		})
	}

	return c.JSON(utils.BaseResponse{
		Success: true,
		Message: "Token berhasil diperbarui",
		Data: dto.RefreshTokenResponse{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		},
	})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Akses ditolak",
			Error: &utils.ErrorResponse{Code: fiber.StatusUnauthorized, Details: "User ID tidak ditemukan dari token"},
		})
	}
	userIDFloat, ok := userIDLocals.(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
			Success: false, Message: "Tipe data User ID tidak valid di dalam token",
		})
	}
	userID := uint(userIDFloat)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Request body tidak valid",
			Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: err.Error()},
		})
	}

	// <-- Ganti `fiber.NewError` dengan response JSON yang konsisten
	if err := h.authService.ChangePassword(uint(userID), req); err != nil {
		if errors.Is(err, services.ErrInvalidCurrentPassword) {
			return c.Status(fiber.StatusForbidden).JSON(utils.BaseResponse{
				Success: false, Message: "Password lama yang Anda masukkan salah",
				Error: &utils.ErrorResponse{Code: fiber.StatusForbidden, Details: err.Error()},
			})
		}
		if errors.Is(err, services.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(utils.BaseResponse{
				Success: false, Message: "Pengguna tidak ditemukan",
				Error: &utils.ErrorResponse{Code: fiber.StatusNotFound, Details: err.Error()},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
			Success: false, Message: "Gagal memperbarui password",
			Error: &utils.ErrorResponse{Code: fiber.StatusInternalServerError, Details: err.Error()},
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Password berhasil diperbarui",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// (Fungsi ini tidak perlu diubah karena sudah menangani error secara umum)
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Header Authorization tidak ditemukan",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Format token tidak valid",
		})
	}

	err := h.authService.Logout(tokenString)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
			Success: false, Message: "Gagal melakukan logout",
			Error: &utils.ErrorResponse{Code: fiber.StatusInternalServerError, Details: err.Error()},
		})
	}

	return c.JSON(utils.BaseResponse{
		Success: true,
		Message: "Logout berhasil. Mohon hapus token di sisi client.",
	})
}
