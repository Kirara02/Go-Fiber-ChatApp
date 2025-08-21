package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"main/config"
	"main/dto"

	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client    *openai.Client
	modelName string
}

// NewOpenAIService membuat instance baru dari OpenAIService.
func NewOpenAIService(cfg *config.Config) (AIService, error) {
	if cfg.OpenaiAPIKey == "" {
		return nil, errors.New("kunci API OpenAI tidak ditemukan di konfigurasi")
	}
	client := openai.NewClient(cfg.OpenaiAPIKey)
	modelName := cfg.GeminiModel
	if modelName == "" {
		// Fallback ke model default jika tidak ada yang diatur di config.
		modelName = openai.GPT4oMini
		log.Println("PERINGATAN: OPENAI_MODEL tidak diatur, menggunakan fallback ke gpt-4o-mini.")
	}
	return &OpenAIService{client: client, modelName: modelName}, nil
}

// summarize adalah fungsi internal untuk membuat ringkasan menggunakan OpenAI.
func (s *OpenAIService) Summarize(ctx context.Context, oldSummary, userMessage, aiResponse string) (string, error) {
	prompt := fmt.Sprintf(
		"Ringkaslah percakapan berikut dengan sangat singkat dalam satu atau dua kalimat. Pertahankan poin-poin kunci.\n\n"+
			"Ringkasan Sebelumnya:\n%s\n\n"+
			"Interaksi Terbaru:\nUser: %s\nAI: %s\n\n"+
			"Ringkasan Baru:", oldSummary, userMessage, aiResponse,
	)

	req := openai.ChatCompletionRequest{
		Model: s.modelName, // Bisa gunakan model yang lebih cepat/murah untuk meringkas
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: prompt},
		},
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("gagal meringkas: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("gagal mengekstrak teks ringkasan dari OpenAI")
}

// GenerateContent (Non-Streaming) dengan logika summarization.
func (s *OpenAIService) GenerateContent(ctx context.Context, previousSummary string, newUserMessage string) (*dto.AIChatResponse, error) {
	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang"

	// --- PERUBAHAN DI SINI ---
	contextualPrompt := fmt.Sprintf(
		"Konteks percakapan sejauh ini: \"%s\"\n\n"+
			"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token:\nUser: %s", previousSummary, newUserMessage,
	)

	// Dapatkan respons utama dari AI
	req := openai.ChatCompletionRequest{
		Model: s.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: contextualPrompt},
		},
		// MaxTokens dihapus untuk mengandalkan instruksi di prompt.
	}
	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("openai tidak memberikan respons utama yang valid")
	}
	aiMainResponse := resp.Choices[0].Message.Content

	// Buat ringkasan baru
	newSummary, err := s.Summarize(context.Background(), previousSummary, newUserMessage, aiMainResponse)
	if err != nil {
		log.Printf("Peringatan: Gagal memperbarui ringkasan OpenAI, menggunakan ringkasan lama. Error: %v", err)
		newSummary = previousSummary // Fallback jika gagal
	}

	// Kembalikan semuanya
	return &dto.AIChatResponse{
		Response:   aiMainResponse,
		Provider:   "openai",
		Model:      s.modelName,
		NewSummary: newSummary,
	}, nil
}

func (s *OpenAIService) GenerateContentStream(ctx context.Context, previousSummary, newUserMessage string) (streamChan <-chan string, fullResponseChan <-chan string, err error) {
	const systemPrompt = "anda adalah seorang asisten Ai yang ramah tergantung mood, seorang profesional developer, dan sangat antusias dengan budaya otaku jepang"

	// --- PERUBAHAN DI SINI ---
	contextualPrompt := fmt.Sprintf(
		"Konteks percakapan sejauh ini: \"%s\"\n\n"+
			"Dengan konteks tersebut, jawablah pertanyaan user berikut dengan ringkas, max response nya 3000 token:\nUser: %s", previousSummary, newUserMessage,
	)

	req := openai.ChatCompletionRequest{
		Model: s.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: contextualPrompt},
		},
		Stream: true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuat stream OpenAI: %w", err)
	}

	sChan := make(chan string)
	fChan := make(chan string, 1) // Channel dengan buffer 1

	go func() {
		defer close(sChan)
		defer close(fChan)
		defer stream.Close()

		var fullResponse string
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Printf("Streaming error dari OpenAI: %v", err)
				break
			}
			chunk := response.Choices[0].Delta.Content
			fullResponse += chunk
			sChan <- chunk
		}
		fChan <- fullResponse
	}()

	return sChan, fChan, nil
}
