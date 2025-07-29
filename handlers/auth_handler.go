package handlers

import (
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

func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				return fmt.Sprintf("Field %s wajib diisi", e.Field())
			case "email":
				return fmt.Sprintf("Field %s harus berupa format email yang valid", e.Field())
			case "min":
				return fmt.Sprintf("Field %s harus memiliki minimal %s karakter", e.Field(), e.Param())
			default:
				return fmt.Sprintf("Field %s tidak valid", e.Field())
			}
		}
	}
	return err.Error()
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
			
			Error: &utils.ErrorResponse{
				Code:    fiber.StatusBadRequest,
				Details: formatValidationError(err),
			},
		})
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if err.Error() == "email already registered" {
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
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Kredensial tidak valid",
			Error: &utils.ErrorResponse{Code: fiber.StatusUnauthorized, Details: "Email atau password salah"},
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
	req := new(dto.RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Cannot parse JSON", Error: &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: err.Error()},
		})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.BaseResponse{
			Success: false, Message: "Refresh token wajib diisi",
			Error:   &utils.ErrorResponse{Code: fiber.StatusBadRequest, Details: formatValidationError(err)},
		})
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
			Success: false, Message: "Gagal memperbarui token",
			Error:   &utils.ErrorResponse{Code: fiber.StatusUnauthorized, Details: err.Error()},
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
		return fiber.NewError(fiber.StatusUnauthorized, "User ID tidak ditemukan")
	}
	userID, _ := userIDLocals.(float64)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}
	
	// Panggil service yang baru
	if err := h.authService.ChangePassword(uint(userID), req); err != nil {
		// Kembalikan status 'Forbidden' jika password lama salah
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Password berhasil diperbarui",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
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
			Error:   &utils.ErrorResponse{Code: fiber.StatusInternalServerError, Details: err.Error()},
		})
	}

	return c.JSON(utils.BaseResponse{
		Success: true,
		Message: "Logout berhasil. Mohon hapus token di sisi client.",
	})
}