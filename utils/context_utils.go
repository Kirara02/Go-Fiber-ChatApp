package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ExtractUserIDFromContext adalah fungsi helper untuk mengambil userID
// yang telah disimpan oleh middleware otentikasi dari context Fiber.
// Nama fungsi dibuat lebih deskriptif karena sekarang bersifat publik.
func ExtractUserIDFromContext(c *fiber.Ctx) (uint, error) {
	idLocals := c.Locals("user_id")
	if idLocals == nil {
		// Mengembalikan error yang jelas yang bisa ditangani oleh handler.
		return 0, fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token (tidak ada di context)")
	}

	// JWT claims seringkali mendekode angka sebagai float64, jadi kita tangani kasus itu.
	idFloat, ok := idLocals.(float64)
	if !ok {
		// Ini adalah kasus error yang tidak terduga, kemungkinan besar karena kesalahan konfigurasi.
		return 0, fiber.NewError(fiber.StatusInternalServerError, "Tipe data ID pengguna tidak valid di context")
	}

	return uint(idFloat), nil
}