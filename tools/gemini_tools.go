package tools

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

// GeminiToolDef sekarang hanya mendefinisikan deklarasi fungsi untuk AI.
type GeminiToolDef struct {
	FunctionDeclaration *genai.FunctionDeclaration
}

// AvailableGeminiTools adalah peta deklarasi fungsi yang tersedia untuk AI.
// Isinya tetap sama, hanya definisinya yang terlihat lebih bersih.
var AvailableGeminiTools = map[string]GeminiToolDef{
	"getTimeIndonesia": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name:        "getTimeIndonesia",
			Description: "Mendapatkan waktu dan tanggal saat ini. Ini adalah zona waktu utama (WIB/Jakarta). Gunakan tool ini jika pengguna bertanya tentang waktu secara umum tanpa menyebutkan negara atau kota lain.",
		},
	},
	"getCurrentTimeInTokyo": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name:        "getCurrentTimeInTokyo",
			Description: "Mendapatkan waktu dan tanggal spesifik untuk Tokyo, Jepang. Hanya gunakan tool ini jika pengguna secara eksplisit menyebutkan 'Tokyo' atau 'Jepang'.",
		},
	},
	"getProductPrice": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name:        "getProductPrice",
			Description: "Mendapatkan harga dari sebuah produk spesifik berdasarkan namanya.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"productName": {
						Type:        genai.TypeString,
						Description: "Nama produk yang ingin dicari harganya, e.g., 'Nendoroid Hatsune Miku', 'Figma Snow Miku'.",
					},
				},
				Required: []string{"productName"},
			},
		},
	},
	"listAvailableProducts": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name:        "listAvailableProducts",
			Description: "Mendapatkan daftar semua nama produk yang harganya tersedia di dalam sistem.",
		},
	},
	"displayProductCatalog": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name: "displayProductCatalog",
			Description: "Menampilkan katalog produk yang lengkap kepada pengguna. Gunakan ini sebagai langkah terakhir setelah semua data produk dan harga terkumpul.",
			Parameters: &genai.Schema{
				Type:        genai.TypeObject,
				Description: "Objek yang berisi katalog produk dan format tampilan.",
				Properties: map[string]*genai.Schema{
					// --- TAMBAHKAN PARAMETER BARU INI ---
					"format": {
						Type:        genai.TypeString,
						Description: "Format output yang diinginkan. Gunakan 'json' HANYA jika pengguna secara eksplisit meminta output dalam format JSON. Jika tidak, gunakan 'text'.",
						Enum:        []string{"text", "json"}, // Enum membatasi pilihan AI
					},
					"products": {
						Type:        genai.TypeArray,
						Description: "Daftar objek produk, masing-masing berisi nama dan harga.",
						Items: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"name": {
									Type: genai.TypeString, Description: "Nama lengkap produk.",
								},
								"price": {
									Type: genai.TypeString, Description: "Harga produk dalam Rupiah.",
								},
							},
						},
					},
				},
				Required: []string{"format", "products"}, // format sekarang wajib diisi
			},
		},
	},
	"filterProductsByPrice": {
		FunctionDeclaration: &genai.FunctionDeclaration{
			Name: "filterProductsByPrice",
			Description: "Menyaring dan mendapatkan daftar nama produk yang harganya lebih rendah atau lebih tinggi dari nilai tertentu. Sangat berguna untuk pertanyaan seperti 'produk apa saja di bawah 1 juta?' atau 'berikan item yang lebih mahal dari 500 ribu'.",
			Parameters: &genai.Schema{
				// ... (parameter sama seperti countProductsByPrice)
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"comparison": {
						Type: genai.TypeString,
						Description: "Jenis perbandingan. Harus 'less_than' atau 'greater_than'.",
					},
					"price": {
						Type: genai.TypeInteger,
						Description: "Nilai harga untuk dibandingkan, dalam bentuk angka.",
					},
				},
				Required: []string{"comparison", "price"},
			},
		},
	},
}

// GetAvailableGeminiTools mengembalikan tool dalam format yang dibutuhkan oleh genai.Tool.
func GetAvailableGeminiTools() *genai.Tool {
	var declarations []*genai.FunctionDeclaration
	for _, toolDef := range AvailableGeminiTools {
		declarations = append(declarations, toolDef.FunctionDeclaration)
	}
	return &genai.Tool{FunctionDeclarations: declarations}
}

// ExecuteToolByName adalah "dispatcher" yang dipanggil oleh service.
// Ia sekarang memanggil fungsi-fungsi publik dari file lain.
func ExecuteToolByName(name, args string, products []ProductDTO) (string, error) {
	switch name {
	case "getProductPrice":
		return ExecuteGetProductPrice(args, products)
	case "listAvailableProducts":
		return ExecuteListAvailableProducts(args, products)
	case "displayProductCatalog":
		return ExecuteDisplayProductCatalog(args)
	case "filterProductsByPrice":
		return ExecuteFilterProductsByPrice(args, products)
	case "getCurrentTimeInTokyo":
		return ExecuteGetCurrentTimeInTokyo(args)
	case "getTimeIndonesia":
		return ExecuteGetTimeIndonesia(args)
	default:
		return "", fmt.Errorf("tool tidak dikenal: %s", name)
	}
}
