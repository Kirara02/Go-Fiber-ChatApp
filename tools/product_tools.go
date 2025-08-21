package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// ProductDTO adalah Data Transfer Object untuk produk dari API.
type ProductDTO struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Stock int    `json:"stock"`
}

func parsePrice(priceStr string) (int, error) {
	// 1. Hapus "Rp" dan spasi
	noCurrency := strings.TrimSpace(strings.TrimPrefix(priceStr, "Rp"))
	// 2. Hapus pemisah ribuan (titik)
	noSeparators := strings.ReplaceAll(noCurrency, ".", "")
	// 3. Hapus pemisah ribuan (koma) jika ada
	noSeparators = strings.ReplaceAll(noSeparators, ",", "")

	// 4. Konversi ke integer
	return strconv.Atoi(noSeparators)
}

// FetchProductsFromAPI membuat panggilan HTTP GET ke endpoint produk.
// Dibuat menjadi publik (Fetch...)
func FetchProductsFromAPI() ([]ProductDTO, error) {
	apiURL := "http://localhost:8080/api/products"
	log.Printf("INFO: Menghubungi API produk di %s", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("ERROR: Gagal menghubungi API produk: %v", err)
		return nil, fmt.Errorf("server produk tidak dapat dijangkau")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: API produk mengembalikan status non-OK: %s", resp.Status)
		return nil, fmt.Errorf("server produk mengembalikan error: %s", resp.Status)
	}

	var products []ProductDTO
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		log.Printf("ERROR: Gagal mengurai JSON dari API produk: %v", err)
		return nil, fmt.Errorf("format data dari server produk tidak valid")
	}

	log.Printf("INFO: Berhasil mengambil %d produk dari API", len(products))
	return products, nil
}

func ExecuteGetProductPrice(args string, products []ProductDTO) (string, error) {
	var params struct {
		ProductName string `json:"productName"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("argumen JSON tidak valid: %w", err)
	}

	if products == nil {
		return "Maaf, data produk tidak tersedia saat ini.", nil
	}

	searchQuery := strings.ToLower(params.ProductName)
	for _, product := range products {
		if strings.Contains(strings.ToLower(product.Name), searchQuery) {
			return product.Price, nil
		}
	}
	return fmt.Sprintf("Maaf, produk '%s' tidak ditemukan.", params.ProductName), nil
}

func ExecuteListAvailableProducts(_ string, products []ProductDTO) (string, error) {
	if products == nil {
		return "Maaf, data produk tidak tersedia saat ini.", nil
	}
	if len(products) == 0 {
		return "Saat ini tidak ada produk yang terdaftar.", nil
	}

	var productNames []string
	for _, product := range products {
		productNames = append(productNames, product.Name)
	}
	return strings.Join(productNames, ", "), nil
}

func ExecuteDisplayProductCatalog(args string) (string, error) {
	var params struct {
		Format   string       `json:"format"`
		Products []ProductDTO `json:"products"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("argumen JSON tidak valid: %w", err)
	}

	if len(params.Products) == 0 {
		return "Tidak ada produk untuk ditampilkan.", nil
	}

	// Fungsi ini hanya akan dipanggil untuk format 'text', jadi kita format sebagai markdown.
	var builder strings.Builder
	builder.WriteString("Ini dia daftar harganya:\n\n")
	for _, p := range params.Products {
		builder.WriteString(fmt.Sprintf("*   **%s:** %s\n", p.Name, p.Price))
	}
	
	return builder.String(), nil
}

func ExecuteFilterProductsByPrice(args string, products []ProductDTO) (string, error) {
	var params struct {
		Comparison string `json:"comparison"` // "less_than" atau "greater_than"
		Price      int    `json:"price"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("argumen JSON tidak valid: %w", err)
	}

	if products == nil {
		return "Maaf, data produk tidak tersedia.", nil
	}

	var matchingProducts []string
	for _, product := range products {
		price, err := parsePrice(product.Price)
		if err != nil {
			continue
		}

		if params.Comparison == "less_than" && price < params.Price {
			matchingProducts = append(matchingProducts, product.Name)
		} else if params.Comparison == "greater_than" && price > params.Price {
			matchingProducts = append(matchingProducts, product.Name)
		}
	}

	if len(matchingProducts) == 0 {
		return "Tidak ada produk yang cocok dengan kriteria tersebut.", nil
	}

	// Kembalikan daftar nama produk yang dipisahkan koma
	return strings.Join(matchingProducts, ", "), nil
}
