package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for real-time updates
func (h *Handler) WebSocketHandler(c *gin.Context) {
	conn, err := h.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	h.wsClients[conn] = true
	defer delete(h.wsClients, conn)

	// Send initial connection message
	conn.WriteJSON(gin.H{
		"type":      "connected",
		"timestamp": time.Now().Unix(),
	})

	// Keep connection alive with ping/pong
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	// Read messages from client (for ping/pong and commands)
	go func() {
		defer close(done)
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Handle different message types
			if messageType == websocket.TextMessage {
				h.handleWebSocketMessage(conn, message)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleBroadcast sends messages to all connected WebSocket clients
func (h *Handler) handleBroadcast() {
	for {
		message := <-h.broadcast
		
		// Send to all connected clients
		for client := range h.wsClients {
			err := client.WriteJSON(message)
			if err != nil {
				client.Close()
				delete(h.wsClients, client)
			}
		}
	}
}

// handleWebSocketMessage processes incoming WebSocket messages
func (h *Handler) handleWebSocketMessage(conn *websocket.Conn, message []byte) {
	// Parse message to determine action
	var msg struct {
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "Invalid message format",
		})
		return
	}

	switch msg.Type {
	case "subscribe":
		// Handle subscription to specific events
		h.handleSubscription(conn, msg.Payload)
	case "ping":
		// Respond with pong
		conn.WriteJSON(gin.H{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
		})
	case "get_status":
		// Send current status
		h.sendCurrentStatus(conn)
	default:
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "Unknown message type",
		})
	}
}

// handleSubscription manages event subscriptions for WebSocket clients
func (h *Handler) handleSubscription(conn *websocket.Conn, payload map[string]interface{}) {
	// TODO: Implement subscription management
	// For now, all clients receive all broadcasts
	
	conn.WriteJSON(gin.H{
		"type":    "subscribed",
		"events":  []string{"all"},
		"message": "Subscribed to all events",
	})
}

// sendCurrentStatus sends the current system status to a WebSocket client
func (h *Handler) sendCurrentStatus(conn *websocket.Conn) {
	status := gin.H{
		"type":      "status_update",
		"timestamp": time.Now().Unix(),
		"relayers": gin.H{
			"hermes": h.checkHermesStatus(),
			"rly":    h.checkRlyStatus(),
		},
	}

	conn.WriteJSON(status)
}

// BroadcastEvent sends an event to all connected WebSocket clients
func (h *Handler) BroadcastEvent(eventType string, data interface{}) {
	h.broadcast <- gin.H{
		"type":      eventType,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}
}