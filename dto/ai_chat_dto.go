package dto


// AIChatRequest adalah struktur untuk body request ke endpoint AI chat
type AIChatRequest struct {
	Message string `json:"message"`
}

// AIChatResponse adalah struktur untuk data dalam response dari AI chat
type AIChatResponse struct {
	Response string `json:"response"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}