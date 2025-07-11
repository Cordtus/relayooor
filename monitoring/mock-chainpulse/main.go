package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Metrics that chainpulse will provide
	ibcStuckPackets = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ibc_stuck_packets",
			Help: "Number of stuck IBC packets",
		},
		[]string{"src_chain", "dst_chain", "channel"},
	)

	ibcEffectedPackets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ibc_effected_packets",
			Help: "Total number of successfully relayed packets",
		},
		[]string{"src_chain", "dst_chain", "channel", "chain_id"},
	)

	ibcUneffectedPackets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ibc_uneffected_packets",
			Help: "Total number of failed packet relay attempts",
		},
		[]string{"src_chain", "dst_chain", "channel", "chain_id"},
	)

	ibcFrontrunCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ibc_frontrun_counter",
			Help: "Number of frontrunning events",
		},
		[]string{"chain_id", "frontrunned_by"},
	)

	ibcHandshakeStates = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ibc_handshake_states",
			Help: "IBC channel handshake states",
		},
		[]string{"chain_id", "channel", "counterparty_channel", "port", "state"},
	)
)

func init() {
	prometheus.MustRegister(ibcStuckPackets)
	prometheus.MustRegister(ibcEffectedPackets)
	prometheus.MustRegister(ibcUneffectedPackets)
	prometheus.MustRegister(ibcFrontrunCounter)
	prometheus.MustRegister(ibcHandshakeStates)
}

type ChannelPair struct {
	SrcChain    string
	DstChain    string
	SrcChannel  string
	DstChannel  string
}

var monitoredChannels = []ChannelPair{
	{
		SrcChain:   "cosmoshub-4",
		DstChain:   "osmosis-1",
		SrcChannel: "channel-141",
		DstChannel: "channel-0",
	},
	{
		SrcChain:   "osmosis-1",
		DstChain:   "cosmoshub-4",
		SrcChannel: "channel-0",
		DstChannel: "channel-141",
	},
}

func simulateMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update stuck packets (should be low or zero in healthy state)
			for _, ch := range monitoredChannels {
				stuck := float64(rand.Intn(3)) // 0-2 stuck packets
				ibcStuckPackets.WithLabelValues(ch.SrcChain, ch.DstChain, ch.SrcChannel).Set(stuck)
			}

			// Simulate packet flow
			for _, ch := range monitoredChannels {
				// Most packets succeed
				effected := float64(rand.Intn(50) + 10)
				uneffected := float64(rand.Intn(5))
				
				ibcEffectedPackets.WithLabelValues(ch.SrcChain, ch.DstChain, ch.SrcChannel, ch.SrcChain).Add(effected)
				ibcUneffectedPackets.WithLabelValues(ch.SrcChain, ch.DstChain, ch.SrcChannel, ch.SrcChain).Add(uneffected)
			}

			// Simulate occasional frontrunning
			if rand.Float32() < 0.3 {
				chains := []string{"cosmoshub-4", "osmosis-1"}
				relayers := []string{"relayer-1", "relayer-2", "relayer-3"}
				
				chain := chains[rand.Intn(len(chains))]
				relayer := relayers[rand.Intn(len(relayers))]
				
				ibcFrontrunCounter.WithLabelValues(chain, relayer).Inc()
			}

			// Update channel states (should be OPEN)
			for _, ch := range monitoredChannels {
				ibcHandshakeStates.WithLabelValues(
					ch.SrcChain,
					ch.SrcChannel,
					ch.DstChannel,
					"transfer",
					"OPEN",
				).Set(3) // 3 = OPEN state
			}
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	// Simulate authentication check
	rpcUsername := os.Getenv("RPC_USERNAME")
	rpcPassword := os.Getenv("RPC_PASSWORD")
	
	if rpcUsername != "" && rpcPassword != "" {
		log.Printf("Mock chainpulse started with RPC authentication enabled")
	} else {
		log.Printf("Mock chainpulse started without RPC authentication")
	}

	// Start metrics simulation
	go simulateMetrics()

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler)
	r.Handle("/metrics", promhttp.Handler())

	port := "3001"
	log.Printf("Mock chainpulse listening on :%s", port)
	log.Printf("This is a POC mimicking real chainpulse metrics for Cosmos Hub <> Osmosis")
	
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		log.Fatal(err)
	}
}