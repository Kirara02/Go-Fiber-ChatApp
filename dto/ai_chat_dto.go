package dto

// ChatMessage tidak berubah.
type ChatMessage struct {
	Role    string `json:"role"` // "user" atau "assistant"
	Content string `json:"content"`
}

// AIChatRequest sekarang memiliki field opsional untuk system prompt.
// Field ini hanya akan berpengaruh pada pesan pertama dalam sebuah sesi.
type AIChatRequest struct {
	SessionID    string `json:"session_id"`
	Message      string `json:"message"`
}

// AIChatResponse tidak berubah.
type AIChatResponse struct {
	Response   string `json:"response"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	NewSummary string `json:"newSummary,omitempty"`
}
