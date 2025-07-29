package websocket

import (
	"log"
	"main/repository"
	"sync"
)

type Hub struct {
	rooms map[string]*Room

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) GetOrCreateRoom(roomID string, chatRepo repository.ChatRepository) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		return room
	}

	room := NewRoom(roomID, chatRepo, h)
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
