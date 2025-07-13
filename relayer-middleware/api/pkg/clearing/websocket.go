package clearing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// WebSocketManager manages WebSocket connections for real-time updates
type WebSocketManager struct {
	clients      map[string]map[*Client]bool
	clientsMutex sync.RWMutex
	redis        *redis.Client
	logger       *zap.Logger
	pubsub       *redis.PubSub
	broadcast    chan BroadcastMessage
	register     chan *Client
	unregister   chan *Client
}

// Client represents a WebSocket client
type Client struct {
	conn         *websocket.Conn
	send         chan []byte
	topics       map[string]bool
	topicsMutex  sync.RWMutex
	manager      *WebSocketManager
	id           string
	walletAddr   string
	pingTicker   *time.Ticker
	lastActivity time.Time
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	Topic   string
	Message WebSocketMessage
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Token     string                 `json:"token,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// SubscriptionMessage for topic subscriptions
type SubscriptionMessage struct {
	Action string   `json:"action"` // subscribe/unsubscribe
	Topics []string `json:"topics"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	maxConnections = 1000 // Maximum concurrent connections
)

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(redis *redis.Client, logger *zap.Logger) *WebSocketManager {
	manager := &WebSocketManager{
		clients:    make(map[string]map[*Client]bool),
		redis:      redis,
		logger:     logger.With(zap.String("component", "websocket")),
		broadcast:  make(chan BroadcastMessage, 256),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
	}

	// Start the manager
	go manager.run()

	// Subscribe to Redis pub/sub for distributed updates
	go manager.subscribeToPubSub()

	return manager
}

// run handles client registration, unregistration, and message broadcasting
func (m *WebSocketManager) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case message := <-m.broadcast:
			m.broadcastToTopic(message.Topic, message.Message)

		case <-ticker.C:
			// Clean up inactive connections
			m.cleanupInactiveClients()
		}
	}
}

// registerClient registers a new client
func (m *WebSocketManager) registerClient(client *Client) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check connection limit
	totalConnections := 0
	for _, clients := range m.clients {
		totalConnections += len(clients)
	}

	if totalConnections >= maxConnections {
		m.logger.Warn("Connection limit reached", zap.Int("total", totalConnections))
		client.conn.Close()
		return
	}

	// Register client for each topic
	client.topicsMutex.RLock()
	for topic := range client.topics {
		if m.clients[topic] == nil {
			m.clients[topic] = make(map[*Client]bool)
		}
		m.clients[topic][client] = true
	}
	client.topicsMutex.RUnlock()

	m.logger.Info("Client registered",
		zap.String("client_id", client.id),
		zap.String("wallet", maskWallet(client.walletAddr)),
		zap.Int("topics", len(client.topics)),
	)
}

// unregisterClient removes a client
func (m *WebSocketManager) unregisterClient(client *Client) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Remove from all topics
	client.topicsMutex.RLock()
	for topic := range client.topics {
		if clients, ok := m.clients[topic]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(m.clients, topic)
			}
		}
	}
	client.topicsMutex.RUnlock()

	// Close channels
	close(client.send)
	client.pingTicker.Stop()

	m.logger.Info("Client unregistered",
		zap.String("client_id", client.id),
		zap.String("wallet", maskWallet(client.walletAddr)),
	)
}

// broadcastToTopic sends a message to all clients subscribed to a topic
func (m *WebSocketManager) broadcastToTopic(topic string, message WebSocketMessage) {
	m.clientsMutex.RLock()
	clients := m.clients[topic]
	m.clientsMutex.RUnlock()

	if len(clients) == 0 {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		m.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	m.logger.Debug("Broadcasting to topic",
		zap.String("topic", topic),
		zap.Int("clients", len(clients)),
		zap.String("type", message.Type),
	)

	// Send to all clients subscribed to this topic
	for client := range clients {
		select {
		case client.send <- data:
			// Message sent successfully
		default:
			// Client's send channel is full, skip
			m.logger.Warn("Client send buffer full",
				zap.String("client_id", client.id),
				zap.String("topic", topic),
			)
		}
	}

	// Also publish to Redis for other instances
	m.publishToRedis(topic, message)
}

// subscribeToPubSub subscribes to Redis pub/sub for distributed updates
func (m *WebSocketManager) subscribeToPubSub() {
	ctx := context.Background()
	m.pubsub = m.redis.Subscribe(ctx, "clearing:updates:*")

	ch := m.pubsub.Channel()
	for msg := range ch {
		// Parse channel name to get topic
		topic := msg.Channel[len("clearing:updates:"):]

		var wsMsg WebSocketMessage
		if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
			m.logger.Error("Failed to unmarshal pub/sub message", zap.Error(err))
			continue
		}

		// Broadcast to local clients
		m.broadcastToTopic(topic, wsMsg)
	}
}

// publishToRedis publishes a message to Redis for other instances
func (m *WebSocketManager) publishToRedis(topic string, message WebSocketMessage) {
	ctx := context.Background()
	channel := fmt.Sprintf("clearing:updates:%s", topic)

	data, err := json.Marshal(message)
	if err != nil {
		m.logger.Error("Failed to marshal for Redis", zap.Error(err))
		return
	}

	if err := m.redis.Publish(ctx, channel, data).Err(); err != nil {
		m.logger.Error("Failed to publish to Redis", zap.Error(err))
	}
}

// cleanupInactiveClients removes clients that haven't sent a pong recently
func (m *WebSocketManager) cleanupInactiveClients() {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	cutoff := time.Now().Add(-pongWait)
	removed := 0

	for topic, clients := range m.clients {
		for client := range clients {
			if client.lastActivity.Before(cutoff) {
				delete(clients, client)
				removed++
				go func(c *Client) {
					c.conn.Close()
				}(client)
			}
		}
		if len(clients) == 0 {
			delete(m.clients, topic)
		}
	}

	if removed > 0 {
		m.logger.Info("Cleaned up inactive clients", zap.Int("count", removed))
	}
}

// HandleWebSocket handles WebSocket connections
func (m *WebSocketManager) HandleWebSocket(c *gin.Context) {
	// Get wallet from session (optional)
	wallet := c.GetString("wallet")

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client
	client := &Client{
		conn:         conn,
		send:         make(chan []byte, 256),
		topics:       make(map[string]bool),
		manager:      m,
		id:           uuid.New().String(),
		walletAddr:   wallet,
		pingTicker:   time.NewTicker(pingPeriod),
		lastActivity: time.Now(),
	}

	// Default subscriptions based on authentication
	if wallet != "" {
		// Subscribe to user-specific updates
		client.topics[fmt.Sprintf("user:%s", wallet)] = true
	}

	// Register client
	m.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()

	// Send welcome message
	welcome := WebSocketMessage{
		Type:      "connected",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id": client.id,
			"version":   "1.0",
		},
	}
	if data, err := json.Marshal(welcome); err == nil {
		client.send <- data
	}
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.lastActivity = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.manager.logger.Error("WebSocket read error",
					zap.String("client_id", c.id),
					zap.Error(err),
				)
			}
			break
		}

		c.lastActivity = time.Now()

		// Handle subscription messages
		var subMsg SubscriptionMessage
		if err := json.Unmarshal(message, &subMsg); err == nil {
			c.handleSubscription(subMsg)
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-c.pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscription handles topic subscription/unsubscription
func (c *Client) handleSubscription(msg SubscriptionMessage) {
	c.topicsMutex.Lock()
	defer c.topicsMutex.Unlock()

	switch msg.Action {
	case "subscribe":
		for _, topic := range msg.Topics {
			// Validate topic access
			if c.canAccessTopic(topic) {
				c.topics[topic] = true
				c.manager.logger.Debug("Client subscribed to topic",
					zap.String("client_id", c.id),
					zap.String("topic", topic),
				)
			}
		}

	case "unsubscribe":
		for _, topic := range msg.Topics {
			delete(c.topics, topic)
			c.manager.logger.Debug("Client unsubscribed from topic",
				zap.String("client_id", c.id),
				zap.String("topic", topic),
			)
		}
	}

	// Re-register to update topic subscriptions
	c.manager.register <- c
}

// canAccessTopic checks if client can access a topic
func (c *Client) canAccessTopic(topic string) bool {
	// User-specific topics require authentication
	if strings.HasPrefix(topic, "user:") {
		requiredWallet := strings.TrimPrefix(topic, "user:")
		return c.walletAddr == requiredWallet
	}

	// Token-specific topics are public
	if strings.HasPrefix(topic, "token:") {
		return true
	}

	// Channel-specific topics are public
	if strings.HasPrefix(topic, "channel:") {
		return true
	}

	// Global topics are public
	if topic == "global" || topic == "platform" {
		return true
	}

	return false
}

// Broadcast sends a message to all clients subscribed to a token
func (m *WebSocketManager) Broadcast(token string, message WebSocketMessage) {
	m.broadcast <- BroadcastMessage{
		Topic:   fmt.Sprintf("token:%s", token),
		Message: message,
	}
}

// BroadcastToUser sends a message to a specific user
func (m *WebSocketManager) BroadcastToUser(wallet string, message WebSocketMessage) {
	m.broadcast <- BroadcastMessage{
		Topic:   fmt.Sprintf("user:%s", wallet),
		Message: message,
	}
}

// Subscribe creates a channel for receiving updates on a topic
func (m *WebSocketManager) Subscribe(topic string) <-chan WebSocketMessage {
	ch := make(chan WebSocketMessage, 10)

	// This is a simplified implementation
	// In production, you'd want to properly manage these channels
	go func() {
		// Subscribe to Redis pub/sub
		ctx := context.Background()
		pubsub := m.redis.Subscribe(ctx, fmt.Sprintf("clearing:updates:%s", topic))
		defer pubsub.Close()

		msgCh := pubsub.Channel()
		for msg := range msgCh {
			var wsMsg WebSocketMessage
			if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err == nil {
				select {
				case ch <- wsMsg:
				default:
					// Channel full, skip message
				}
			}
		}
	}()

	return ch
}

// Unsubscribe closes a subscription channel
func (m *WebSocketManager) Unsubscribe(topic string, ch <-chan WebSocketMessage) {
	// In a real implementation, you'd properly clean up the subscription
	// For now, we'll just drain the channel
	go func() {
		for range ch {
			// Drain channel
		}
	}()
}

// GetConnectionStats returns current connection statistics
func (m *WebSocketManager) GetConnectionStats() map[string]interface{} {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	totalConnections := 0
	topicStats := make(map[string]int)

	for topic, clients := range m.clients {
		count := len(clients)
		totalConnections += count
		topicStats[topic] = count
	}

	return map[string]interface{}{
		"total_connections": totalConnections,
		"topics":            topicStats,
		"timestamp":         time.Now(),
	}
}

// Additional imports
import (
	"strings"

	"github.com/google/uuid"
)