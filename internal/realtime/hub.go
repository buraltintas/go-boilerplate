package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go-boilerplate/internal/observability"
)

// Connection represents a WebSocket connection.
// Single responsibility: manage a single WebSocket connection.
type Connection struct {
	ID     string
	UserID string
	ws     *websocket.Conn
	send   chan []byte
	hub    *Hub
}

// Message represents a WebSocket message.
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// NewConnection creates a new WebSocket connection.
func NewConnection(id, userID string, ws *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:     id,
		UserID: userID,
		ws:     ws,
		send:   make(chan []byte, 256),
		hub:    hub,
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
// Runs in its own goroutine.
func (c *Connection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.ws.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				observability.Error("websocket error", "error", err, "connection_id", c.ID)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			observability.Warn("invalid websocket message", "error", err, "connection_id", c.ID)
			continue
		}

		// Handle incoming message (extend this for your use cases)
		observability.Debug("websocket message received", "type", msg.Type, "connection_id", c.ID, "user_id", c.UserID)
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// Runs in its own goroutine.
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				observability.Error("websocket write error", "error", err, "connection_id", c.ID)
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start starts the connection's read and write pumps.
func (c *Connection) Start() {
	go c.writePump()
	go c.readPump()
}

// Hub maintains active WebSocket connections and broadcasts messages.
// Single responsibility: manage all WebSocket connections.
// Design: Simple hub pattern for broadcasting events to connected clients.
type Hub struct {
	connections map[string]*Connection
	broadcast   chan []byte
	register    chan *Connection
	unregister  chan *Connection
	mu          sync.RWMutex
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		broadcast:   make(chan []byte),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}
}

// Run starts the hub's main loop.
// Should be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.ID] = conn
			h.mu.Unlock()
			observability.Info("websocket connection registered", "connection_id", conn.ID, "user_id", conn.UserID)

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.ID]; ok {
				delete(h.connections, conn.ID)
				close(conn.send)
			}
			h.mu.Unlock()
			observability.Info("websocket connection unregistered", "connection_id", conn.ID, "user_id", conn.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.send <- message:
				default:
					// Client can't keep up, close connection
					close(conn.send)
					delete(h.connections, conn.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(messageType string, payload interface{}) error {
	msg := Message{
		Type:    messageType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

// SendToUser sends a message to a specific user's connections.
func (h *Hub) SendToUser(userID, messageType string, payload interface{}) error {
	msg := Message{
		Type:    messageType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conn := range h.connections {
		if conn.UserID == userID {
			select {
			case conn.send <- data:
			default:
				// Skip if send buffer is full
			}
		}
	}

	return nil
}

// GetConnectionCount returns the number of active connections.
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}
