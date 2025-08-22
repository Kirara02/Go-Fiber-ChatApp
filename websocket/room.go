// package websocket

// import (
// 	"encoding/json"
// 	"log"
// 	"main/domain"
// 	"main/repository"
// 	"strconv"
// )

// type BroadcastMessage struct {
// 	Sender  *Client
// 	Payload []byte
// }

// type Room struct {
// 	ID         string
// 	clients    map[*Client]bool
// 	broadcast  chan BroadcastMessage
// 	Register   chan *Client
// 	Unregister chan *Client
// 	chatRepo   repository.ChatRepository
// 	hub        *Hub
// }

// func NewRoom(id string, chatRepo repository.ChatRepository, hub *Hub) *Room {
// 	return &Room{
// 		ID:         id,
// 		clients:    make(map[*Client]bool),
// 		broadcast:  make(chan BroadcastMessage),
// 		Register:   make(chan *Client),
// 		Unregister: make(chan *Client),
// 		chatRepo:   chatRepo,
// 		hub:        hub,
// 	}
// }

// func (r *Room) Run() {
// 	for {
// 		select {
// 		case client := <-r.Register:
// 			r.clients[client] = true
// 			log.Printf("Client %s bergabung ke room %s", client.UserName, r.ID)

// 		case client := <-r.Unregister:
// 			if _, ok := r.clients[client]; ok {
// 				delete(r.clients, client)
// 				close(client.Send)
// 				log.Printf("Client %s meninggalkan room %s", client.UserName, r.ID)
// 				if r.hub != nil {
// 					r.hub.removeRoomIfEmpty(r)
// 				}
// 			}

// 		case message := <-r.broadcast:

// 			var dbMessage domain.ChatMessage

// 			if err := json.Unmarshal(message.Payload, &dbMessage); err != nil {
// 				log.Printf("Error unmarshal pesan untuk DB: %v", err)
// 				continue
// 			}

// 			dbMessage.UserID = message.Sender.UserID

// 			roomIDUint, _ := strconv.ParseUint(r.ID, 10, 32)
// 			dbMessage.RoomID = uint(roomIDUint)

// 			if err := r.chatRepo.CreateMessage(&dbMessage); err != nil {
// 				log.Printf("Gagal menyimpan pesan ke DB: %v", err)
// 			}

// 			log.Printf("Akan menyiarkan pesan ke %d klien di room %s.", len(r.clients), r.ID)

// 			for client := range r.clients {
// 				select {
// 				case client.Send <- message.Payload:
// 					// --- LOGGING TAMBAHAN ---
// 					log.Printf("Pesan berhasil dikirim ke channel client %s.", client.UserName)
// 				default:
// 					log.Printf("GAGAL: Channel client %s penuh. Menutup koneksi.", client.UserName)
// 					close(client.Send)
// 					delete(r.clients, client)
// 				}
// 			}
// 		}
// 	}
// }


package websocket

import (
	"context"
	"encoding/json"
	"log"
	"main/domain"
	"main/dto"
	"main/repository"
	"main/services"
	"strconv"
	"time"
)

type BroadcastMessage struct {
	Sender  *Client
	Payload []byte
}

type AIMessage struct {
	Sender  *Client
	Content string
}

type Room struct {
	ID             string
	clients        map[*Client]bool
	broadcast      chan BroadcastMessage
	Register       chan *Client
	Unregister     chan *Client
	processAI      chan AIMessage
	chatRepo       repository.ChatRepository
	hub            *Hub
	aiService      services.AIService
	sessionService services.SessionService
}

func NewRoom(id string, chatRepo repository.ChatRepository, hub *Hub, aiService services.AIService, sessionService services.SessionService) *Room {
	return &Room{
		ID:             id,
		clients:        make(map[*Client]bool),
		broadcast:      make(chan BroadcastMessage),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		processAI:      make(chan AIMessage),
		chatRepo:       chatRepo,
		hub:            hub,
		aiService:      aiService,
		sessionService: sessionService,
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
			// Simpan pesan ke DB hanya jika dikirim oleh user asli
			if message.Sender != nil {
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
			}

			// Siarkan ke semua klien
			for client := range r.clients {
				select {
				case client.Send <- message.Payload:
				default:
					close(client.Send)
					delete(r.clients, client)
				}
			}

		case aiMsg := <-r.processAI:
			log.Printf("INFO: Memproses permintaan AI di Room %s", r.ID)
			go r.handleAIRequest(aiMsg)
		}
	}
}

func (r *Room) handleAIRequest(aiMsg AIMessage) {
	sessionID := "room-" + r.ID
	roomIDUint, _ := strconv.ParseUint(r.ID, 10, 32)

	session, err := r.sessionService.GetOrCreateSession(sessionID, 0) // user ID 0 untuk sesi AI room
	if err != nil {
		log.Printf("ERROR: Gagal mendapatkan sesi AI untuk room %s: %v", r.ID, err)
		return
	}

	aiResponse, err := r.aiService.GenerateContent(context.Background(), session.Summary, aiMsg.Content)
	if err != nil {
		log.Printf("ERROR: Gagal mendapatkan respons AI: %v", err)
		errorDTO := dto.ChatMessageResponse{
			Type:       "chat",
			SenderID:   0,
			SenderName: "AI Assistant",
			Content:    "Maaf, saya sedang mengalami sedikit kendala teknis.",
			RoomID:     uint(roomIDUint),
			CreatedAt:  time.Now(),
		}
		jsonError, _ := json.Marshal(errorDTO)
		aiMsg.Sender.Send <- jsonError // Kirim error hanya ke pengirim
		return
	}

	r.sessionService.UpdateSummary(session, aiResponse.NewSummary)

	aiMessageDTO := dto.ChatMessageResponse{
		Type:       "chat",
		SenderID:   0,
		SenderName: "AI Assistant",
		Content:    aiResponse.Response,
		RoomID:     uint(roomIDUint),
		CreatedAt:  time.Now(),
	}
	jsonMessage, _ := json.Marshal(aiMessageDTO)

	// Siarkan respons AI ke semua orang di room.
	// Kita gunakan Sender: nil agar pesan ini tidak disimpan lagi ke DB sebagai pesan user.
	r.broadcast <- BroadcastMessage{
		Sender:  nil,
		Payload: jsonMessage,
	}
	log.Printf("INFO: Respons AI berhasil disiarkan ke Room %s.", r.ID)
}
