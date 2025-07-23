package websocket

import (
	"encoding/json"
	"log"
	"main/domain"
	"main/repository"
	"strconv"
)

type BroadcastMessage struct {
	Sender  *Client
	Payload []byte
}

type Room struct {
	ID         string
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	chatRepo   repository.ChatRepository
	hub        *Hub
}

func NewRoom(id string, chatRepo repository.ChatRepository, hub *Hub) *Room {
	return &Room{
		ID:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client), 
		Unregister: make(chan *Client),
		chatRepo:   chatRepo,
		hub:        hub,
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.clients[client] = true
			log.Printf("Client %s bergabung ke room %s", client.UserName, r.ID)

		case client := <-r.Unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
				log.Printf("Client %s meninggalkan room %s", client.UserName, r.ID)
				if r.hub != nil {
					r.hub.removeRoomIfEmpty(r)
				}
			}

		case message := <-r.broadcast:
			
			var dbMessage domain.ChatMessage
			
			if err := json.Unmarshal(message.Payload, &dbMessage); err != nil {
				log.Printf("Error unmarshal pesan untuk DB: %v", err)
				continue
			}

			dbMessage.UserID = message.Sender.UserID
			
			roomIDUint, _ := strconv.ParseUint(r.ID, 10, 32)
			dbMessage.RoomID = uint(roomIDUint)
			
			if err := r.chatRepo.CreateMessage(&dbMessage); err != nil {
				log.Printf("Gagal menyimpan pesan ke DB: %v", err)
			}

			log.Printf("Akan menyiarkan pesan ke %d klien di room %s.", len(r.clients), r.ID)

			for client := range r.clients {
				select {
					case client.Send <- message.Payload:
					// --- LOGGING TAMBAHAN ---
					log.Printf("Pesan berhasil dikirim ke channel client %s.", client.UserName)
				default:
					log.Printf("GAGAL: Channel client %s penuh. Menutup koneksi.", client.UserName)
					close(client.Send)
					delete(r.clients, client)
				}
			}
		}
	}
}