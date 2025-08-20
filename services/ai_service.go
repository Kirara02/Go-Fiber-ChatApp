package services

import (
	"context"
	"main/dto"
)

type AIService interface {
	GenerateContent(ctx context.Context, prompt string) (*dto.AIChatResponse, error)
}
