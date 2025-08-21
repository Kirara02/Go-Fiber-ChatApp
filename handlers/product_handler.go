package handlers

import (
	"log"
	"main/dto"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var productData = []dto.Product{
	{
		ID:    1,
		Name:  "Nendoroid Hatsune Miku",
		Price: "Rp 750.000",
		Stock: 15,
	},
	{
		ID:    2,
		Name:  "Figma Snow Miku",
		Price: "Rp 1.200.000",
		Stock: 8,
	},
	{
		ID:    3,
		Name:  "Hatsune Miku V4X",
		Price: "Rp 2.500.000",
		Stock: 20,
	},
	{
		ID:    4,
		Name:  "1/7 Scale Figure Hatsune Miku (Symphony 2023 Ver.)",
		Price: "Rp 3.200.000",
		Stock: 5,
	},
	{
		ID:    5,
		Name:  "Nendoroid Kagamine Rin",
		Price: "Rp 720.000",
		Stock: 12,
	},
	{
		ID:    6,
		Name:  "Nendoroid Kagamine Len",
		Price: "Rp 720.000",
		Stock: 11,
	},
	{
		ID:    7,
		Name:  "POP UP PARADE Megurine Luka",
		Price: "Rp 650.000",
		Stock: 18,
	},
}

// ProductHandler sekarang menjadi struct kosong karena tidak memiliki dependency.
type ProductHandler struct{}

// NewProductHandler sekarang tidak memerlukan argumen apa pun.
func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// GetAllProducts sekarang mengambil data langsung dari variabel `productData`.
func (h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	log.Println("INFO: Endpoint /api/products dipanggil")

	// Langsung kembalikan data dummy sebagai respons JSON.
	// Tidak ada panggilan service atau error handling yang kompleks yang diperlukan.
	return c.Status(http.StatusOK).JSON(productData)
}