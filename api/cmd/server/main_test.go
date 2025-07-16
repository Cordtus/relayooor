package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test health endpoint
func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var result HealthResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Status)
}

// Test metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var result MetricsResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.StuckPackets, 0)
	assert.GreaterOrEqual(t, result.ActiveChannels, 0)
}

// Test user transfers with valid wallet
func TestUserTransfersValidWallet(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/user/osmo1test123/transfers", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var transfers []UserTransfer
	err := json.NewDecoder(resp.Body).Decode(&transfers)
	require.NoError(t, err)
	assert.IsType(t, []UserTransfer{}, transfers)
}

// Test user transfers with invalid wallet
func TestUserTransfersInvalidWallet(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/user/invalid/transfers", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	
	var result map[string]string
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "Invalid wallet address", result["error"])
}

// Test stuck packets endpoint
func TestStuckPacketsEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/packets/stuck", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var packets []StuckPacket
	err := json.NewDecoder(resp.Body).Decode(&packets)
	require.NoError(t, err)
	assert.IsType(t, []StuckPacket{}, packets)
}

// Test clear packets with valid request
func TestClearPacketsValid(t *testing.T) {
	router := setupTestRouter()
	
	clearReq := ClearPacketRequest{
		PacketIDs: []string{"packet-1", "packet-2"},
		Wallet:    "osmo1test123",
		Signature: "valid-signature",
	}
	
	body, _ := json.Marshal(clearReq)
	req, _ := http.NewRequest("POST", "/api/packets/clear", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var result ClearPacketResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.TxHash)
	assert.Equal(t, clearReq.PacketIDs, result.Cleared)
}

// Test clear packets with invalid signature
func TestClearPacketsInvalidSignature(t *testing.T) {
	router := setupTestRouter()
	
	clearReq := ClearPacketRequest{
		PacketIDs: []string{"packet-1"},
		Wallet:    "osmo1test123",
		Signature: "", // Invalid empty signature
	}
	
	body, _ := json.Marshal(clearReq)
	req, _ := http.NewRequest("POST", "/api/packets/clear", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	
	var result map[string]string
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "Invalid signature", result["error"])
}

// Test clear packets with invalid body
func TestClearPacketsInvalidBody(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("POST", "/api/packets/clear", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// Test channels endpoint
func TestChannelsEndpoint(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/channels", nil)
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusOK, resp.Code)
	
	var channels []Channel
	err := json.NewDecoder(resp.Body).Decode(&channels)
	require.NoError(t, err)
	assert.Greater(t, len(channels), 0)
	assert.Equal(t, "channel-0", channels[0].ChannelID)
}

// Test CORS headers
func TestCORSHeaders(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("OPTIONS", "/api/metrics", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp := httptest.NewRecorder()
	
	router.ServeHTTP(resp, req)
	
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Contains(t, resp.Header().Get("Access-Control-Allow-Origin"), "localhost")
	assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "GET")
}

// Helper function to setup test router
func setupTestRouter() *mux.Router {
	r := mux.NewRouter()
	
	// Register all routes
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}).Methods("GET")
	
	api := r.PathPrefix("/api").Subrouter()
	
	// Add all API routes here (simplified for testing)
	api.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := MetricsResponse{
			StuckPackets:   0,
			ActiveChannels: 2,
			PacketFlowRate: 12.5,
			SuccessRate:    98.5,
		}
		json.NewEncoder(w).Encode(metrics)
	}).Methods("GET")
	
	api.HandleFunc("/channels", func(w http.ResponseWriter, r *http.Request) {
		channels := []Channel{
			{
				ChannelID:             "channel-0",
				CounterpartyChannelID: "channel-141",
				SourceChain:           "osmosis-1",
				DestinationChain:      "cosmoshub-4",
				State:                 "OPEN",
				PendingPackets:        0,
				TotalPackets:          15234,
			},
		}
		json.NewEncoder(w).Encode(channels)
	}).Methods("GET")
	
	api.HandleFunc("/packets/stuck", func(w http.ResponseWriter, r *http.Request) {
		packets := []StuckPacket{}
		json.NewEncoder(w).Encode(packets)
	}).Methods("GET")
	
	api.HandleFunc("/user/{wallet}/transfers", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		wallet := vars["wallet"]
		
		if !isValidWalletAddress(wallet) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid wallet address"})
			return
		}
		
		transfers := []UserTransfer{}
		json.NewEncoder(w).Encode(transfers)
	}).Methods("GET")
	
	api.HandleFunc("/packets/clear", func(w http.ResponseWriter, r *http.Request) {
		var req ClearPacketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		
		if !verifyWalletSignature(req.Wallet, req.Signature) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid signature"})
			return
		}
		
		resp := ClearPacketResponse{
			Status:  "success",
			TxHash:  "TEST_TX_" + generateTxHash(),
			Cleared: req.PacketIDs,
			Failed:  []string{},
			Message: "Packets cleared successfully",
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}).Methods("POST")
	
	// Add CORS headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	return r
}

// Benchmark tests
func BenchmarkHealthEndpoint(b *testing.B) {
	router := setupTestRouter()
	
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
}

func BenchmarkUserTransfers(b *testing.B) {
	router := setupTestRouter()
	
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/user/osmo1test123/transfers", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
}