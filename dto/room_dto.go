package dto

import (
	"main/domain"
	"time"
)

type CreateRoomRequest struct {
	Name    string `json:"name"`
	UserIDs []uint `json:"userIds"`
}

type RoomResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	IsPrivate bool           `json:"is_private"`
	IsGroup   bool           `json:"is_group"`
	OwnerID   *uint          `json:"owner_id,omitempty"`
	Users     []UserResponse `json:"users,omitempty"`
	LastMessage    string         `json:"last_message,omitempty"`
	LastMessageAt  *time.Time     `json:"last_message_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// ToRoomResponse mengonversi domain.Room menjadi DTO.
func ToRoomResponse(room *domain.Room,  currentUserID uint, includeMembers bool) RoomResponse {
	isGroup := len(room.Users) > 2
	displayName := room.Name

	if !isGroup && len(room.Users) == 2 {
		for _, user := range room.Users {
			if user.ID != currentUserID {
				displayName = user.Name 
				break
			}
		}
	}

	var userResponses []UserResponse
	if includeMembers {
		// Hanya konversi data user jika diminta
		userResponses = ToUserResponses(room.Users)
	}
	
	var lastMessageContent string
	var lastMessageAt *time.Time

	if room.LastMessage.ID != 0 {
		lastMessageContent = room.LastMessage.Content
		lastMessageAt = &room.LastMessage.CreatedAt
	}

	return RoomResponse{
		ID:        room.ID,
		Name:      displayName,
		IsPrivate: room.IsPrivate,
		IsGroup:   isGroup,
		OwnerID:   room.OwnerID,
		Users:     userResponses, 
		LastMessage:    lastMessageContent,
		LastMessageAt:  lastMessageAt,  
		CreatedAt: room.CreatedAt,
	}
}

func ToRoomResponses(rooms []*domain.Room, currentUserID uint, includeMembers bool) []RoomResponse {
	var responses []RoomResponse
	for _, room := range rooms {
		responses = append(responses, ToRoomResponse(room, currentUserID, includeMembers))
	}
	return responses
}