package dto

import (
	"main/domain"
	"time"
)

type ChatMessageResponse struct {
	Type       string    `json:"type"`
	SenderID   uint      `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	RoomID     uint      `json:"room_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChatHistoryResponse struct {
	Type     string                `json:"type"`
	Messages []ChatMessageResponse `json:"messages"`
}

func ToChatMessageResponse(msg *domain.ChatMessage) ChatMessageResponse {
	var senderName string
	if msg.User.Name != "" {
		senderName = msg.User.Name
	}

	return ChatMessageResponse{
		Type:       "chat",
		SenderID:   msg.UserID,
		SenderName: senderName,
		Content:    msg.Content,
		RoomID:     msg.RoomID,
		CreatedAt:  msg.CreatedAt,
	}
}