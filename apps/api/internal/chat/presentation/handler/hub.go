package handler

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ChatClient struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

type ChatHub struct {
	mu      sync.RWMutex
	clients map[string][]*ChatClient
}

var DefaultChatHub = NewChatHub()

func NewChatHub() *ChatHub {
	return &ChatHub{
		clients: make(map[string][]*ChatClient),
	}
}

func (h *ChatHub) Register(userID string, conn *websocket.Conn) *ChatClient {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := &ChatClient{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	h.clients[userID] = append(h.clients[userID], client)

	// Start write pump for this connection
	go client.writePump()

	return client
}

func (h *ChatHub) Unregister(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.clients[client.UserID]
	if !ok {
		return
	}

	for i, c := range clients {
		if c == client {
			close(c.Send)
			h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.clients[client.UserID]) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *ChatHub) SendToUser(userID string, eventType string, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		log.Printf("failed to marshal chat payload: %v", err)
		return
	}

	h.mu.RLock()
	clients, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for _, c := range clients {
		select {
		case c.Send <- payload:
		default:
			// Client's channel buffer is full, unregister asynchronously
			go h.Unregister(c)
		}
	}
}

func (c *ChatClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-ticker.C:
			// Send a ping to keep connection alive
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
