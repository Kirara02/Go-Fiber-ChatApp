package services

import (
	"context"
	"errors"
	"log"
	"main/config"
	"main/dto"

	"github.com/google/generative-ai-go/genai"
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

	modelName := "gemini-2.5-flash"
	model := client.GenerativeModel(modelName)

	return &GeminiService{
		client:    model,
		modelName: modelName,
	}, nil
}

func (s *GeminiService) GenerateContent(ctx context.Context, prompt string) (*dto.AIChatResponse, error) {
	resp, err := s.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	var aiMessage string
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				aiMessage += string(txt)
			}
		}
	} else {
		return nil, errors.New("gemini tidak memberikan respons yang valid")
	}

	return &dto.AIChatResponse{
		Response: aiMessage,
		Provider: "google",
		Model:    s.modelName,
	}, nil
}