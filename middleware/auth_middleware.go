package middleware

import (
	"main/config"
	"main/repository"
	"main/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config, tokenRepo repository.TokenRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string
		authHeader := c.Get("Authorization")

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
				Success: false, Message: "Token not provided",
			})
		}

		isInvalidated, err := tokenRepo.IsTokenInvalidated(tokenString)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(utils.BaseResponse{
				Success: false, Message: "Gagal memverifikasi token", Error: &utils.ErrorResponse{Details: err.Error()},
			})
		}
		if isInvalidated {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
				Success: false, Message: "Token invalid",
			})
		}

		// Proses validasi JWT berjalan seperti biasa.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Metode signing tidak terduga")
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || claims["type"] != "access" {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Tipe token tidak valid")
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.BaseResponse{
				Success: false, Message: "Token tidak valid atau kedaluwarsa", Error: &utils.ErrorResponse{Details: err.Error()},
			})
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Locals("user_id", claims["user_id"])
		c.Locals("user_name", claims["name"])

		return c.Next()
	}
}
