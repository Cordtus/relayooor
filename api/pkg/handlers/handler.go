package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	hermesURL    string
	redisClient  *redis.Client
	wsUpgrader   websocket.Upgrader
	wsClients    map[*websocket.Conn]bool
	broadcast    chan interface{}
}

func NewHandler() *Handler {
	hermesURL := os.Getenv("HERMES_REST_URL")
	if hermesURL == "" {
		hermesURL = "http://localhost:5185"
	}

	// Initialize Redis client
	redisOpts := &redis.Options{
		Addr:     "redis:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
	redisClient := redis.NewClient(redisOpts)

	h := &Handler{
		hermesURL:   hermesURL,
		redisClient: redisClient,
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure CORS for WebSocket
				return true
			},
		},
		wsClients: make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}),
	}

	// Start WebSocket broadcaster
	go h.handleBroadcast()

	return h
}

// Health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	// Check Redis connection
	ctx := context.Background()
	_, err := h.redisClient.Ping(ctx).Result()
	redisStatus := "healthy"
	if err != nil {
		redisStatus = "unhealthy"
	}

	// Check Hermes REST API
	hermesStatus := "unhealthy"
	resp, err := http.Get(fmt.Sprintf("%s/version", h.hermesURL))
	if err == nil && resp.StatusCode == http.StatusOK {
		hermesStatus = "healthy"
		resp.Body.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": gin.H{
			"redis":  redisStatus,
			"hermes": hermesStatus,
		},
		"timestamp": time.Now().Unix(),
	})
}

// Helper function to call Hermes REST API
func (h *Handler) callHermesAPI(endpoint string) (interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s", h.hermesURL, endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper function to send POST request to Hermes
func (h *Handler) postHermesAPI(endpoint string, data interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s%s", h.hermesURL, endpoint),
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}