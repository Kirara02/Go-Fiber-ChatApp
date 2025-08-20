package services

import (
	"context"
	"errors"
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

func (s *OpenAIService) GenerateContentStream(ctx context.Context, prompt string) (<-chan string, error) {
	req := openai.ChatCompletionRequest{
		Model: s.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Stream: true, // <-- PENTING: Aktifkan mode streaming
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	streamChan := make(chan string)

	go func() {
		defer close(streamChan)
		defer stream.Close()
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Printf("Streaming error from OpenAI: %v", err)
				break
			}
			streamChan <- response.Choices[0].Delta.Content
		}
	}()

	return streamChan, nil
}
