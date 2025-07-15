# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Create a simple Go server
RUN cat > main.go << 'EOF'
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
)

func enableCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "service": "relayooor-api",
    })
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "totalPackets": 1234567,
        "successRate": 98.5,
        "activeRelayers": 42,
        "totalVolume": 15000000,
        "stuckPackets": 23,
        "avgClearingTime": 45,
    })
}

func handleStuckPackets(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "packets": []interface{}{},
        "total": 0,
    })
}

func handleChannels(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "channels": []interface{}{},
        "total": 0,
    })
}

func handleRelayers(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "relayers": []interface{}{},
        "total": 0,
    })
}

func main() {
    http.HandleFunc("/health", handleHealth)
    http.HandleFunc("/api/metrics", handleMetrics)
    http.HandleFunc("/api/packets/stuck", handleStuckPackets)
    http.HandleFunc("/api/channels", handleChannels)
    http.HandleFunc("/api/relayers", handleRelayers)
    
    port := os.Getenv("API_PORT")
    if port == "" {
        port = "3000"
    }
    
    log.Printf("Starting API server on port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
EOF

# Build the binary
RUN go build -o server main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 3000

# Run the binary
CMD ["./server"]