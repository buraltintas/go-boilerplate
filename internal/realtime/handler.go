package realtime

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go-boilerplate/internal/contextutil"
	"go-boilerplate/internal/observability"
	"go-boilerplate/internal/response"
	"go-boilerplate/internal/security"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin should be configured based on CORS policy
	CheckOrigin: func(r *http.Request) bool {
		// In production, check against allowed origins
		return true
	},
}

// Handler handles WebSocket connections.
// Single responsibility: WebSocket HTTP upgrade and connection setup.
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// Connect handles WebSocket upgrade requests.
// GET /ws
// Requires authentication middleware.
func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID, err := contextutil.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		observability.Error("failed to upgrade websocket", "error", err, "user_id", userID)
		return
	}

	// Generate connection ID
	connID, err := security.GenerateSecureToken(16)
	if err != nil {
		observability.Error("failed to generate connection ID", "error", err)
		ws.Close()
		return
	}

	// Create new connection
	conn := NewConnection(connID, userID, ws, h.hub)

	// Register connection with hub
	h.hub.register <- conn

	// Start connection pumps
	conn.Start()
}
