package services

import (
	"errors"
	"log"
	"main/config"
)

// ProvideAIService is a Wire provider function.
// It reads the configuration and returns the appropriate implementation of AIService.
func ProvideAIService(cfg *config.Config) (AIService, error) {
	switch cfg.AIProvider {
	case "google":
		log.Println("Initializing AI Provider: Google Gemini")
		// NewGeminiService now returns (AIService, error)
		return NewGeminiService(cfg)
	case "openai":
		log.Println("Initializing AI Provider: OpenAI ChatGPT")
		// NewOpenAIService returns AIService, so we wrap it to match the signature
		return NewOpenAIService(cfg)
	default:
		// Return an error if the provider is not supported
		return nil, errors.New("invalid AI_PROVIDER specified: " + cfg.AIProvider)
	}
}
