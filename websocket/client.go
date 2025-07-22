package websocket

import (
	"encoding/json"
	"log"
	"main/domain"

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
		
		// Buat object pesan untuk di-marshal ke JSON
		chatMessage := domain.ChatMessage{
			Type:       "chat",
			SenderID:   c.UserID,
			SenderName: c.UserName,
			Content:    string(message),
		}
		
		jsonMessage, err := json.Marshal(chatMessage)
		if err != nil {
			log.Println("Error marshal chat message:", err)
			continue
		}
		
		c.room.broadcast <- BroadcastMessage{
			Sender:  c, 
			Payload: jsonMessage,
		}
	}
}

func (c *Client) WritePump() {
	for message := range c.Send {
		if err := c.conn.WriteMessage(ws.TextMessage, message); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}