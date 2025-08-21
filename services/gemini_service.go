package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"main/config"
	"main/dto"

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

func (s *GeminiService) GenerateContent(ctx context.Context, previousSummary, newUserMessage string) (*dto.AIChatResponse, error) {
	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang"

	// --- PERUBAHAN DI SINI ---
	// Menambahkan instruksi panjang respons langsung ke dalam prompt.
	contextualPrompt := fmt.Sprintf(
		"PERAN ANDA: %s\n\n"+
			"Konteks percakapan sejauh ini: \"%s\"\n\n"+
			"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token:\nUser: %s", systemPrompt, previousSummary, newUserMessage,
	)

	// Menghapus batas token teknis untuk mengandalkan instruksi di prompt.
	s.client.GenerationConfig = genai.GenerationConfig{}

	resp, err := s.client.GenerateContent(ctx, genai.Text(contextualPrompt))
	if err != nil {
		return nil, err
	}

	aiMainResponse, err := extractTextFromResponse(resp)
	if err != nil {
		return nil, err
	}

	newSummary, err := s.Summarize(context.Background(), previousSummary, newUserMessage, aiMainResponse)
	if err != nil {
		log.Printf("Peringatan: Gagal memperbarui ringkasan, menggunakan ringkasan lama. Error: %v", err)
		newSummary = previousSummary
	}

	return &dto.AIChatResponse{
		Response:   aiMainResponse,
		Provider:   "google",
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
