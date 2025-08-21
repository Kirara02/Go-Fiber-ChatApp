package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"main/config"
	"main/dto"
	"main/tools"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client    *genai.GenerativeModel
	modelName string
}

func NewGeminiService(cfg *config.Config) (AIService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		log.Printf("Gagal membuat client Gemini: %v", err)
		return nil, errors.New("gagal menginisialisasi layanan Gemini")
	}
	modelName := cfg.GeminiModel
	if modelName == "" {
		// Fallback ke model default jika tidak ada yang diatur di config.
		modelName = "gemini-2.5-flash"
		log.Println("PERINGATAN: GEMINI_MODEL tidak diatur, menggunakan fallback ke gpt-4o-mini.")
	}

	model := client.GenerativeModel(modelName)
	model.Tools = []*genai.Tool{tools.GetAvailableGeminiTools()}

	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockOnlyHigh},
	}

	return &GeminiService{client: model, modelName: modelName}, nil
}

// Summarize sekarang menjadi bagian dari interface AIService.
func (s *GeminiService) Summarize(ctx context.Context, oldSummary, userMessage, aiResponse string) (string, error) {
	prompt := fmt.Sprintf(
		"Ringkaslah percakapan berikut dengan sangat singkat dalam satu atau dua kalimat. Pertahankan poin-poin kunci.\n\n"+
			"Ringkasan Sebelumnya:\n%s\n\n"+
			"Interaksi Terbaru:\nUser: %s\nAI: %s\n\n"+
			"Ringkasan Baru:", oldSummary, userMessage, aiResponse,
	)

	resp, err := s.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gagal memanggil API summarize: %w", err)
	}

	return extractTextFromResponse(resp)
}

func extractTextFromResponse(resp *genai.GenerateContentResponse) (string, error) {
	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("DEBUG: Respons mentah dari Gemini:\n%s", string(jsonData))

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		finishReason := "UNKNOWN"
		if len(resp.Candidates) > 0 {
			finishReason = resp.Candidates[0].FinishReason.String()
		}
		return "", fmt.Errorf("respons AI tidak valid atau diblokir (FinishReason: %s)", finishReason)
	}

	var textContent string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			textContent += string(txt)
		}
	}

	if textContent == "" {
		return "", errors.New("gagal mengekstrak konten teks dari respons AI")
	}

	return textContent, nil
}

// func (s *GeminiService) GenerateContent(ctx context.Context, previousSummary, newUserMessage string) (*dto.AIChatResponse, error) {
// 	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang. Zona waktu utama Anda adalah Waktu Indonesia Barat (WIB). Anda memiliki akses ke alat (tools) untuk membantu menjawab pertanyaan, seperti mencari harga produk. Gunakan alat tersebut jika relevan untuk menjawab permintaan pengguna."

// 	chatSession := s.client.StartChat()
// 	chatSession.History = []*genai.Content{
// 		{
// 			Role: "user",
// 			Parts: []genai.Part{genai.Text(fmt.Sprintf(
// 				"PERAN ANDA: %s\n\n"+
// 					"Konteks percakapan sejauh ini: \"%s\"\n\n"+
// 					"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token.",
// 				systemPrompt, previousSummary,
// 			))},
// 		},
// 		{
// 			Role:  "model",
// 			Parts: []genai.Part{genai.Text("Baik, saya mengerti peran, konteks, dan alat yang saya miliki. Saya siap menerima pertanyaan.")},
// 		},
// 	}

// 	const maxTurns = 5
// 	var finalResponseText strings.Builder
// 	currentParts := []genai.Part{genai.Text(newUserMessage)}

// 	// Cache ini hanya akan hidup selama satu kali permintaan GenerateContent.
// 	var productCache []tools.ProductDTO
// 	isCachePopulated := false // Flag untuk memastikan API hanya dipanggil sekali

// 	for i := 0; i < maxTurns; i++ {
// 		log.Printf("INFO: Memulai putaran percakapan tool ke-%d", i+1)

// 		resp, err := chatSession.SendMessage(ctx, currentParts...)
// 		if err != nil {
// 			return nil, fmt.Errorf("gagal pada panggilan SendMessage di putaran %d: %w", i+1, err)
// 		}

// 		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
// 			return nil, fmt.Errorf("menerima respons kosong dari Gemini di putaran %d", i+1)
// 		}

// 		var functionCalls []genai.FunctionCall
// 		for _, part := range resp.Candidates[0].Content.Parts {
// 			if text, ok := part.(genai.Text); ok {
// 				finalResponseText.WriteString(string(text))
// 			}
// 			if fc, ok := part.(genai.FunctionCall); ok {
// 				functionCalls = append(functionCalls, fc)
// 			}
// 		}

// 		if len(functionCalls) == 0 {
// 			log.Println("INFO: Tidak ada lagi pemanggilan tool. Mengakhiri loop.")
// 			break
// 		}

// 		// Implementasi Caching: Panggil API HANYA SEKALI jika diperlukan.
// 		if !isCachePopulated {
// 			needsProductData := false
// 			for _, fc := range functionCalls {
// 				if fc.Name == "listAvailableProducts" || fc.Name == "getProductPrice" {
// 					needsProductData = true
// 					break
// 				}
// 			}

// 			if needsProductData {
// 				log.Println("INFO: Cache produk kosong, mengambil data dari API...")
// 				cachedData, err := tools.FetchProductsFromAPI()
// 				if err != nil {
// 					log.Printf("ERROR: Gagal mengambil data untuk cache: %v", err)
// 				} else {
// 					productCache = cachedData
// 				}
// 			}
// 			isCachePopulated = true // Tandai bahwa kita sudah mencoba mengisi cache
// 		} else {
// 			log.Println("INFO: Menggunakan data produk dari cache.")
// 		}

// 		log.Printf("INFO: Mendeteksi %d pemanggilan tool di putaran %d. Mengeksekusi...", len(functionCalls), i+1)
// 		var toolResults []genai.Part
// 		for _, fc := range functionCalls {
// 			argsBytes, _ := json.Marshal(fc.Args)

// 			// Panggil dispatcher pusat, teruskan cache produk.
// 			result, err := tools.ExecuteToolByName(fc.Name, string(argsBytes), productCache)
// 			if err != nil {
// 				log.Printf("ERROR: Gagal mengeksekusi tool %s: %v", fc.Name, err)
// 				result = fmt.Sprintf("Error: %v", err) // Kirim error kembali ke AI
// 			}

// 			toolResults = append(toolResults, &genai.FunctionResponse{
// 				Name:     fc.Name,
// 				Response: map[string]any{"result": result},
// 			})
// 		}

// 		currentParts = toolResults
// 	}

// 	aiMainResponse := finalResponseText.String()
// 	if aiMainResponse == "" {
// 		return nil, errors.New("setelah beberapa putaran, AI tidak memberikan respons teks akhir")
// 	}

// 	newSummary, err := s.Summarize(context.Background(), previousSummary, newUserMessage, aiMainResponse)
// 	if err != nil {
// 		log.Printf("Peringatan: Gagal memperbarui ringkasan, menggunakan ringkasan lama. Error: %v", err)
// 		newSummary = previousSummary
// 	}

// 	return &dto.AIChatResponse{
// 		Response:   aiMainResponse,
// 		Provider:   "google",
// 		Model:      s.modelName,
// 		NewSummary: newSummary,
// 	}, nil
// }

func (s *GeminiService) GenerateContent(ctx context.Context, previousSummary, newUserMessage string) (*dto.AIChatResponse, error) {
	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang. Zona waktu utama Anda adalah Waktu Indonesia Barat (WIB). Anda memiliki akses ke alat (tools) untuk membantu menjawab pertanyaan. Gunakan alat tersebut jika relevan untuk menjawab permintaan pengguna."

	chatSession := s.client.StartChat()
	chatSession.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{genai.Text(fmt.Sprintf(
				"PERAN ANDA: %s\n\n"+
					"Konteks percakapan sejauh ini: \"%s\"\n\n"+
					"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token.",
				systemPrompt, previousSummary,
			))},
		},
		{
			Role:  "model",
			Parts: []genai.Part{genai.Text("Baik, saya mengerti peran, konteks, dan alat yang saya miliki. Saya siap menerima pertanyaan.")},
		},
	}

	const maxTurns = 5
	currentParts := []genai.Part{genai.Text(newUserMessage)}
	var finalResponseText strings.Builder // Kita gunakan lagi untuk mengakumulasi teks

	var productCache []tools.ProductDTO
	isCachePopulated := false

	for i := 0; i < maxTurns; i++ {
		log.Printf("INFO: Memulai putaran percakapan tool ke-%d", i+1)

		resp, err := chatSession.SendMessage(ctx, currentParts...)
		if err != nil {
			return nil, fmt.Errorf("gagal pada panggilan SendMessage di putaran %d: %w", i+1, err)
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			return nil, fmt.Errorf("menerima respons kosong dari Gemini di putaran %d", i+1)
		}

		var functionCalls []genai.FunctionCall
		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				finalResponseText.WriteString(string(text))
			}
			if fc, ok := part.(genai.FunctionCall); ok {
				functionCalls = append(functionCalls, fc)
			}
		}

		// --- LOGIKA DINAMIS BARU ---
		var nonDisplayTools []genai.FunctionCall
		isJSONRequest := false
		var jsonDataToReturn string

		for _, fc := range functionCalls {
			if fc.Name == "displayProductCatalog" {
				args := fc.Args
				if format, ok := args["format"].(string); ok && format == "json" {
					// Tandai bahwa ini adalah permintaan JSON dan simpan datanya
					isJSONRequest = true
					productsData := map[string]interface{}{"products": args["products"]}
					jsonOutput, err := json.Marshal(productsData)
					if err != nil {
						return nil, fmt.Errorf("gagal me-marshal argumen final: %w", err)
					}
					jsonDataToReturn = string(jsonOutput)
				} else {
					// Jika formatnya 'text', anggap sebagai tool biasa
					nonDisplayTools = append(nonDisplayTools, fc)
				}
			} else {
				// Tambahkan tool lain ke daftar eksekusi
				nonDisplayTools = append(nonDisplayTools, fc)
			}
		}

		// Jika permintaan JSON terdeteksi, langsung hentikan dan kembalikan.
		if isJSONRequest {
			log.Println("INFO: Panggilan 'displayProductCatalog' dengan format 'json' terdeteksi. Mengembalikan JSON murni.")
			newSummary, _ := s.Summarize(context.Background(), previousSummary, newUserMessage, "AI menampilkan katalog produk dalam format JSON.")
			return &dto.AIChatResponse{
				Response:   jsonDataToReturn,
				Provider:   s.modelName,
				Model:      s.modelName,
				NewSummary: newSummary,
			}, nil
		}
		// ------------------------------------

		// Ganti functionCalls dengan daftar yang sudah difilter
		functionCalls = nonDisplayTools

		if len(functionCalls) == 0 {
			log.Println("INFO: Tidak ada lagi pemanggilan tool. Mengakhiri loop.")
			break
		}

		// ... (Logika caching tetap sama, pastikan tool baru ada di daftar pengecekan) ...
		if !isCachePopulated {
			needsProductData := false
			for _, fc := range functionCalls {
				if fc.Name == "listAvailableProducts" || fc.Name == "getProductPrice" || fc.Name == "listAllProductsWithPrices" || fc.Name == "filterProductsByPrice" || fc.Name == "displayProductCatalog" {
					needsProductData = true
					break
				}
			}
			if needsProductData {
				log.Println("INFO: Cache produk kosong, mengambil data dari API...")
				cachedData, err := tools.FetchProductsFromAPI()
				if err != nil {
					log.Printf("ERROR: Gagal mengambil data untuk cache: %v", err)
				} else {
					productCache = cachedData
				}
			}
			isCachePopulated = true
		} else {
			log.Println("INFO: Menggunakan data produk dari cache.")
		}

		var toolResults []genai.Part
		for _, fc := range functionCalls {
			argsBytes, _ := json.Marshal(fc.Args)
			result, err := tools.ExecuteToolByName(fc.Name, string(argsBytes), productCache)
			if err != nil {
				log.Printf("ERROR: Gagal mengeksekusi tool %s: %v", fc.Name, err)
				result = fmt.Sprintf("Error: %v", err)
			}
			toolResults = append(toolResults, &genai.FunctionResponse{Name: fc.Name, Response: map[string]any{"result": result}})
		}
		currentParts = toolResults
	}

	aiMainResponse := finalResponseText.String()
	if aiMainResponse == "" {
		return nil, errors.New("AI tidak menghasilkan respons akhir yang valid setelah beberapa putaran")
	}

	newSummary, err := s.Summarize(context.Background(), previousSummary, newUserMessage, aiMainResponse)
	if err != nil {
		log.Printf("Peringatan: Gagal memperbarui ringkasan, menggunakan ringkasan lama. Error: %v", err)
		newSummary = previousSummary
	}

	return &dto.AIChatResponse{
		Response:   aiMainResponse,
		Provider:   s.modelName,
		Model:      s.modelName,
		NewSummary: newSummary,
	}, nil
}

// GenerateContentStream sekarang diimplementasikan sepenuhnya.
func (s *GeminiService) GenerateContentStream(ctx context.Context, previousSummary, newUserMessage string) (streamChan <-chan string, fullResponseChan <-chan string, err error) {
	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang"

	contextualPrompt := fmt.Sprintf(
		"PERAN ANDA: %s\n\n"+
			"Konteks percakapan sejauh ini: \"%s\"\n\n"+
			"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token:\nUser: %s", systemPrompt, previousSummary, newUserMessage,
	)

	// Memastikan tidak ada batas token teknis.
	s.client.GenerationConfig = genai.GenerationConfig{}

	iter := s.client.GenerateContentStream(ctx, genai.Text(contextualPrompt))

	sChan := make(chan string)
	fChan := make(chan string, 1) // Channel dengan buffer 1

	go func() {
		defer close(sChan)
		defer close(fChan)

		var fullResponse string
		for {
			resp, err := iter.Next()

			jsonData, _ := json.MarshalIndent(resp, "", "  ")
			log.Printf("DEBUG: Chunk mentah dari Gemini Stream:\n%s", string(jsonData))

			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Streaming error dari Gemini: %v", err)
				break
			}

			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						chunk := string(txt)
						fullResponse += chunk
						sChan <- chunk
					}
				}
			}
		}
		fChan <- fullResponse
	}()

	return sChan, fChan, nil
}
