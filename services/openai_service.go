package services

import (
	"context"
	"errors"
	"main/config"
	"main/dto"

	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client    *openai.Client
	modelName string
}

// NewOpenAIService membuat instance baru dari OpenAIService.
func NewOpenAIService(cfg *config.Config) AIService {
	client := openai.NewClient(cfg.OpenaiAPIKey)
	modelName := openai.GPT3Dot5Turbo

	return &OpenAIService{
		client:    client,
		modelName: modelName,
	}
}

func (s *OpenAIService) GenerateContent(ctx context.Context, prompt string) (*dto.AIChatResponse, error) {
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) > 0 {
		return &dto.AIChatResponse{
			Response: resp.Choices[0].Message.Content,
			Provider: "openai",
			Model:    s.modelName,
		}, nil
	}

	return nil, errors.New("openai tidak memberikan respons yang valid")
}