package websocket

import (
	"log"
	"main/repository"
	"main/services"
	"sync"
)

type Hub struct {
	rooms          map[string]*Room
	mu             sync.RWMutex
	aiService      services.AIService
	sessionService services.SessionService
}

func NewHub(aiService services.AIService, sessionService services.SessionService) *Hub {
	return &Hub{
		rooms:          make(map[string]*Room),
		aiService:      aiService,
		sessionService: sessionService,
	}
}

func (h *Hub) GetOrCreateRoom(roomID string, chatRepo repository.ChatRepository) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		return room
	}

	// Teruskan AIService saat membuat Room baru
	room := NewRoom(roomID, chatRepo, h, h.aiService, h.sessionService)
	h.rooms[roomID] = room

	go room.Run()

	return room
}

// Anda bisa menambahkan fungsi untuk membersihkan room yang kosong
func (h *Hub) removeRoomIfEmpty(room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(room.clients) == 0 {
		delete(h.rooms, room.ID)
		log.Printf("Room %s dihapus karena kosong.", room.ID)
	}
}
