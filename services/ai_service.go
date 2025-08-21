package services

import (
	"context"
	"main/dto"
)

// AIService mendefinisikan kontrak untuk semua layanan AI.
type AIService interface {
	GenerateContent(ctx context.Context, previousSummary string, newUserMessage string) (*dto.AIChatResponse, error)

	GenerateContentStream(ctx context.Context, previousSummary, newUserMessage string) (streamChan <-chan string, fullResponseChan <-chan string, err error)

	Summarize(ctx context.Context, oldSummary, userMessage, aiResponse string) (string, error)
}
