package dto

import (
	"main/domain"
	"time"
)

// CreateRoomRequest adalah DTO untuk menerima data saat membuat room baru.
type CreateRoomRequest struct {
	Name    string `json:"name"`
	UserIDs []uint `json:"userIds"` // Daftar ID pengguna yang akan diundang
}

// RoomResponse adalah DTO untuk mengirim detail room ke klien.
type RoomResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	IsPrivate bool           `json:"isPrivate"`
	OwnerID   *uint          `json:"ownerId,omitempty"`
	Users     []UserResponse `json:"users"`
	CreatedAt time.Time      `json:"createdAt"`
}

// ToRoomResponse mengonversi domain.Room menjadi DTO.
func ToRoomResponse(room *domain.Room) RoomResponse {
	return RoomResponse{
		ID:        room.ID,
		Name:      room.Name,
		IsPrivate: room.IsPrivate,
		OwnerID:   room.OwnerID,
		Users:     ToUserResponses(room.Users), 
		CreatedAt: room.CreatedAt,
	}
}

func ToRoomResponses(rooms []*domain.Room) []RoomResponse {
	var responses []RoomResponse
	
	if len(rooms) == 0 {
		return responses
	}

	for _, room := range rooms {
		responses = append(responses, ToRoomResponse(room))
	}

	return responses
}