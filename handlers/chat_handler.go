package handlers

import (
	"encoding/json"
	"log"
	"main/dto"
	"main/repository"
	"main/services"
	chat "main/websocket"
	"strconv"

	ws "github.com/gofiber/websocket/v2"
)

type ChatHandler struct {
	hub         *chat.Hub
	userRepo    repository.UserRepository
	chatRepo    repository.ChatRepository
	roomService services.RoomService
}

func NewChatHandler(
	hub *chat.Hub,
	userRepo repository.UserRepository,
	chatRepo repository.ChatRepository,
	roomService services.RoomService,
) *ChatHandler {
	return &ChatHandler{
		hub:         hub,
		userRepo:    userRepo,
		chatRepo:    chatRepo,
		roomService: roomService,
	}
}

func (h *ChatHandler) HandleWebSocket(c *ws.Conn) {
	roomIDStr := c.Params("roomId")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		log.Println("FATAL: Room ID di URL tidak valid.")
		c.Close()
		return
	}

	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		c.Close()
		return
	}
	userID, _ := userIDLocals.(float64)

	isMember, err := h.roomService.IsUserMember(uint(userID), uint(roomID))
	if err != nil || !isMember {
		log.Printf("SECURITY: Pengguna %d mencoba akses room %d tanpa izin.", uint(userID), uint(roomID))
		c.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.ClosePolicyViolation, "Akses ditolak."))
		c.Close()
		return
	}

	user, err := h.userRepo.GetUserByID(uint(userID))
	if err != nil {
		c.Close()
		return
	}

	history, err := h.chatRepo.GetMessagesByRoomID(uint(roomID), 50, 0)
	if err != nil {
		log.Printf("Gagal memuat riwayat chat untuk room %d: %v", roomID, err)
	}

	room := h.hub.GetOrCreateRoom(roomIDStr, h.chatRepo)

	log.Printf("SUCCESS: Pengguna '%s' (ID: %d) terhubung ke room '%s'.", user.Name, user.ID, roomIDStr)

	client := chat.NewClient(c, room, user.ID, user.Name)

	if len(history) > 0 {
		var messageResponses []dto.ChatMessageResponse
		for _, msg := range history {
			messageResponses = append(messageResponses, dto.ToChatMessageResponse(&msg))
		}
		
		historyResponse := dto.ChatHistoryResponse{
			Type:     "history",
			Messages: messageResponses,
		}
		
		jsonHistory, err := json.Marshal(historyResponse)
		if err == nil {
			client.Send <- jsonHistory
		}
	}

	room.Register <- client

	go client.WritePump()
	
	client.ReadPump()
}