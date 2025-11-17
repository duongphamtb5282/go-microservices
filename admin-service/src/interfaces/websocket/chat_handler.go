package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "text", "system", "user_joined", "user_left"
}

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Send     chan *ChatMessage
	Hub      *ChatHub
}

// ChatHub manages all connected clients
type ChatHub struct {
	clients    map[string]*Client
	broadcast  chan *ChatMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *logging.Logger
}

// NewChatHub creates a new chat hub
func NewChatHub(logger *logging.Logger) *ChatHub {
	return &ChatHub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *ChatMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the chat hub
func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()

			h.logger.Info("Client registered", "client_id", client.ID, "username", client.Username)

			// Send user joined notification
			h.broadcast <- &ChatMessage{
				ID:        uuid.New().String(),
				UserID:    client.ID,
				Username:  client.Username,
				Message:   client.Username + " joined the chat",
				Timestamp: time.Now(),
				Type:      "user_joined",
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()

			h.logger.Info("Client unregistered",
				logging.String("client_id", client.ID),
				logging.String("username", client.Username))

			// Send user left notification
			h.broadcast <- &ChatMessage{
				ID:        uuid.New().String(),
				UserID:    client.ID,
				Username:  client.Username,
				Message:   client.Username + " left the chat",
				Timestamp: time.Now(),
				Type:      "user_left",
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, close it
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.Send)
					delete(h.clients, client.ID)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// GetConnectedUsers returns the list of connected users
func (h *ChatHub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for _, client := range h.clients {
		users = append(users, client.Username)
	}
	return users
}

// GetClientCount returns the number of connected clients
func (h *ChatHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error",
					logging.Error(err),
					logging.String("client_id", c.ID))
			}
			break
		}

		// Parse message
		var msg ChatMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			c.Hub.logger.Error("Failed to parse message",
				logging.Error(err),
				logging.String("client_id", c.ID))
			continue
		}

		// Add metadata
		msg.ID = uuid.New().String()
		msg.UserID = c.ID
		msg.Username = c.Username
		msg.Timestamp = time.Now()
		if msg.Type == "" {
			msg.Type = "text"
		}

		// Broadcast message
		c.Hub.broadcast <- &msg
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message as JSON
			if err := c.Conn.WriteJSON(message); err != nil {
				c.Hub.logger.Error("Failed to write message",
					logging.Error(err),
					logging.String("client_id", c.ID))
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ChatHandler handles WebSocket chat connections
type ChatHandler struct {
	hub    *ChatHub
	logger *logging.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(hub *ChatHub, logger *logging.Logger) *ChatHandler {
	return &ChatHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleChat handles WebSocket connections for chat
func (h *ChatHandler) HandleChat(c *gin.Context) {
	// Get username from query parameter or use default
	username := c.Query("username")
	if username == "" {
		username = "Anonymous_" + uuid.New().String()[:8]
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection",
			logging.Error(err))
		return
	}

	// Create client
	client := &Client{
		ID:       uuid.New().String(),
		Username: username,
		Conn:     conn,
		Send:     make(chan *ChatMessage, 256),
		Hub:      h.hub,
	}

	// Register client
	h.hub.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// HandleChatStats returns chat statistics
func (h *ChatHandler) HandleChatStats(c *gin.Context) {
	stats := gin.H{
		"connected_users": h.hub.GetConnectedUsers(),
		"client_count":    h.hub.GetClientCount(),
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, stats)
}
