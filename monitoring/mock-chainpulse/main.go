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
			Name: "ibc_frontrun_total",
			Help: "Total number of frontrunning events",
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

	// Additional metrics that are commonly used in IBC monitoring
	ibcPacketAgeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ibc_packet_age_seconds",
			Help: "Age of unrelayed packets in seconds",
		},
		[]string{"src_chain", "dst_chain", "channel", "sequence"},
	)

	ibcStuckPacketsDetailed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ibc_stuck_packets_detailed",
			Help: "Detailed stuck packet tracking with user info",
		},
		[]string{"src_chain", "dst_chain", "channel", "sequence", "sender", "receiver", "amount", "denom"},
	)

	chainpulseChains = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "chainpulse_chains",
			Help: "The number of chains being monitored",
		},
	)

	chainpulseTxs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chainpulse_txs",
			Help: "The number of txs processed",
		},
		[]string{"chain_id"},
	)

	chainpulsePackets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chainpulse_packets",
			Help: "The number of packets processed",
		},
		[]string{"chain_id"},
	)

	chainpulseReconnects = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chainpulse_reconnects",
			Help: "The number of times we had to reconnect to the WebSocket",
		},
		[]string{"chain_id"},
	)

	chainpulseErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chainpulse_errors",
			Help: "The number of times an error was encountered",
		},
		[]string{"chain_id", "error_type"},
	)
)

func init() {
	prometheus.MustRegister(ibcStuckPackets)
	prometheus.MustRegister(ibcEffectedPackets)
	prometheus.MustRegister(ibcUneffectedPackets)
	prometheus.MustRegister(ibcFrontrunCounter)
	prometheus.MustRegister(ibcHandshakeStates)
	prometheus.MustRegister(ibcPacketAgeSeconds)
	prometheus.MustRegister(ibcStuckPacketsDetailed)
	prometheus.MustRegister(chainpulseChains)
	prometheus.MustRegister(chainpulseTxs)
	prometheus.MustRegister(chainpulsePackets)
	prometheus.MustRegister(chainpulseReconnects)
	prometheus.MustRegister(chainpulseErrors)
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
				log.Printf("Frontrun event: chain=%s, relayer=%s", chain, relayer)
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

			// Update chainpulse system metrics
			chainpulseChains.Set(2) // Monitoring 2 chains
			
			chains := []string{"cosmoshub-4", "osmosis-1"}
			for _, chain := range chains {
				// Simulate transaction count
				txCount := float64(rand.Intn(100) + 50)
				chainpulseTxs.WithLabelValues(chain).Add(txCount)
				
				// Simulate packet count
				packetCount := float64(rand.Intn(20) + 10)
				chainpulsePackets.WithLabelValues(chain).Add(packetCount)
				
				// Occasionally simulate reconnects
				if rand.Float32() < 0.1 {
					chainpulseReconnects.WithLabelValues(chain).Inc()
				}
				
				// Rarely simulate errors
				if rand.Float32() < 0.05 {
					errorTypes := []string{"timeout", "decode_error", "network_error"}
					errorType := errorTypes[rand.Intn(len(errorTypes))]
					chainpulseErrors.WithLabelValues(chain, errorType).Inc()
				}
			}

			// Simulate packet age for stuck packets
			for _, ch := range monitoredChannels {
				// For each stuck packet, simulate age (we'll simulate 0-2 stuck packets)
				stuckCount := rand.Intn(3)
				for i := 0; i < stuckCount; i++ {
					sequence := fmt.Sprintf("%d", rand.Intn(1000)+1)
					age := float64(rand.Intn(3600) + 60) // 1 minute to 1 hour
					ibcPacketAgeSeconds.WithLabelValues(ch.SrcChain, ch.DstChain, ch.SrcChannel, sequence).Set(age)
					
					// Add detailed stuck packet info
					sender := fmt.Sprintf("%s1%s", ch.SrcChain[:4], generateRandomAddress())
					receiver := fmt.Sprintf("%s1%s", ch.DstChain[:4], generateRandomAddress())
					amount := fmt.Sprintf("%d", rand.Intn(10000)+100)
					denom := "uatom"
					if ch.SrcChain == "osmosis-1" {
						denom = "uosmo"
					}
					
					ibcStuckPacketsDetailed.WithLabelValues(
						ch.SrcChain,
						ch.DstChain,
						ch.SrcChannel,
						sequence,
						sender,
						receiver,
						amount,
						denom,
					).Set(1)
				}
			}
		}
	}
}

func generateRandomAddress() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 38)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
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