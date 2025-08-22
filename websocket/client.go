package websocket

import (
	"encoding/json"
	"log"
	"main/dto"
	"strconv"
	"strings"
	"time"

	ws "github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID   uint
	UserName string
	room     *Room
	conn     *ws.Conn
	Send     chan []byte
}

func NewClient(c *ws.Conn, room *Room, userID uint, userName string) *Client {
	return &Client{
		UserID:   userID,
		UserName: userName,
		room:     room,
		conn:     c,
		Send:     make(chan []byte, 256), // Inisialisasi channel Send
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.room.Unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("unexpected read error: %v", err)
			}
			break
		}

		messageContent := string(message)
		roomIDUint, _ := strconv.ParseUint(c.room.ID, 10, 32)

		// --- LOGIKA BARU ---
		// Langkah 1: Selalu siarkan pesan pengguna terlebih dahulu agar tersimpan dan terlihat.
		chatMessageDTO := dto.ChatMessageResponse{
			Type:       "chat",
			SenderID:   c.UserID,
			SenderName: c.UserName,
			Content:    messageContent,
			RoomID:     uint(roomIDUint),
			CreatedAt:  time.Now(),
		}
		jsonMessage, err := json.Marshal(chatMessageDTO)
		if err != nil {
			log.Println("Error marshal DTO:", err)
			continue
		}

		// Kirim pesan pengguna ke broadcast agar semua orang bisa melihatnya dan tersimpan di DB.
		c.room.broadcast <- BroadcastMessage{
			Sender:  c,
			Payload: jsonMessage,
		}

		// Langkah 2: Setelah disiarkan, periksa apakah pesan ini juga untuk AI.
		if strings.HasPrefix(strings.ToLower(messageContent), "@ai") {
			prompt := strings.TrimSpace(strings.TrimPrefix(messageContent, "@ai"))
			if prompt != "" {
				// Kirim prompt yang sudah bersih ke channel AI untuk diproses.
				c.room.processAI <- AIMessage{
					Sender:  c,
					Content: prompt,
				}
			}
		}
	}
}

func (c *Client) WritePump() {
	for message := range c.Send {
		if err := c.conn.WriteMessage(ws.TextMessage, message); err != nil {
			log.Println("write error:", err)
			c.conn.Close()
			break
		}
	}
}
