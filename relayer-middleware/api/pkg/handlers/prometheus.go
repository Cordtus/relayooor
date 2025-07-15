package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/expfmt"
)

type PrometheusMetric struct {
	Name   string                 `json:"name"`
	Help   string                 `json:"help"`
	Type   string                 `json:"type"`
	Values []PrometheusMetricValue `json:"values"`
}

type PrometheusMetricValue struct {
	Labels map[string]string `json:"labels"`
	Value  float64          `json:"value"`
}

// GetChainpulseMetrics fetches and parses metrics from chainpulse
func (h *Handler) GetChainpulseMetrics(c *gin.Context) {
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://chainpulse:3001"
	}

	resp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chainpulse metrics"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read metrics"})
		return
	}

	// Parse Prometheus metrics
	metrics, err := parsePrometheusMetrics(string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metrics"})
		return
	}

	// Extract key IBC metrics
	ibcMetrics := extractIBCMetrics(metrics)
	
	c.JSON(http.StatusOK, ibcMetrics)
}

// GetPacketFlowMetrics returns packet flow statistics
func (h *Handler) GetPacketFlowMetrics(c *gin.Context) {
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://prometheus:9090"
	}

	// Query for packet flow over time
	query := `sum by (chain_id, src_channel, dst_channel) (rate(ibc_effected_packets[5m]))`
	
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, query))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query Prometheus"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStuckPackets returns currently stuck packets
func (h *Handler) GetStuckPackets(c *gin.Context) {
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://chainpulse:3001"
	}

	resp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read metrics"})
		return
	}

	// Parse and extract stuck packet metrics
	metrics, _ := parsePrometheusMetrics(string(body))
	stuckPackets := []gin.H{}

	for _, metric := range metrics {
		if metric.Name == "ibc_stuck_packets" {
			for _, value := range metric.Values {
				if value.Value > 0 {
					stuckPackets = append(stuckPackets, gin.H{
						"src_chain":   value.Labels["src_chain"],
						"dst_chain":   value.Labels["dst_chain"],
						"src_channel": value.Labels["src_channel"],
						"count":       int(value.Value),
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stuck_packets": stuckPackets,
		"total":        len(stuckPackets),
	})
}

// GetRelayerPerformance returns relayer performance metrics
func (h *Handler) GetRelayerPerformance(c *gin.Context) {
	chainpulseURL := os.Getenv("CHAINPULSE_URL")
	if chainpulseURL == "" {
		chainpulseURL = "http://chainpulse:3001"
	}

	resp, err := http.Get(fmt.Sprintf("%s/metrics", chainpulseURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read metrics"})
		return
	}

	metrics, _ := parsePrometheusMetrics(string(body))
	
	// Calculate performance metrics
	performance := gin.H{
		"effected_packets":   countMetricValues(metrics, "ibc_effected_packets"),
		"uneffected_packets": countMetricValues(metrics, "ibc_uneffected_packets"),
		"frontrun_events":    countMetricValues(metrics, "ibc_frontrun_total"),
		"stuck_packets":      countMetricValues(metrics, "ibc_stuck_packets"),
	}

	// Calculate success rate
	effected := performance["effected_packets"].(int)
	uneffected := performance["uneffected_packets"].(int)
	total := effected + uneffected
	
	if total > 0 {
		performance["success_rate"] = float64(effected) / float64(total) * 100
	} else {
		performance["success_rate"] = 0.0
	}

	c.JSON(http.StatusOK, performance)
}

// Helper functions
func parsePrometheusMetrics(data string) ([]PrometheusMetric, error) {
	var metrics []PrometheusMetric
	
	lines := strings.Split(data, "\n")
	currentMetric := &PrometheusMetric{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "# HELP") {
			parts := strings.SplitN(line, " ", 4)
			if len(parts) >= 4 {
				currentMetric.Name = parts[2]
				currentMetric.Help = parts[3]
			}
		} else if strings.HasPrefix(line, "# TYPE") {
			parts := strings.SplitN(line, " ", 4)
			if len(parts) >= 4 {
				currentMetric.Type = parts[3]
			}
		} else if line != "" && !strings.HasPrefix(line, "#") {
			// Parse metric value line
			value, labels := parseMetricLine(line)
			if value != nil {
				currentMetric.Values = append(currentMetric.Values, PrometheusMetricValue{
					Labels: labels,
					Value:  *value,
				})
			}
		} else if line == "" && currentMetric.Name != "" {
			// End of current metric
			metrics = append(metrics, *currentMetric)
			currentMetric = &PrometheusMetric{}
		}
	}
	
	// Add last metric if exists
	if currentMetric.Name != "" {
		metrics = append(metrics, *currentMetric)
	}
	
	return metrics, nil
}

func parseMetricLine(line string) (*float64, map[string]string) {
	// Split metric name from value
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil, nil
	}
	
	// Parse value
	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, nil
	}
	
	// Parse labels
	labels := make(map[string]string)
	nameAndLabels := parts[0]
	
	if strings.Contains(nameAndLabels, "{") {
		// Extract labels
		start := strings.Index(nameAndLabels, "{")
		end := strings.Index(nameAndLabels, "}")
		if start != -1 && end != -1 && end > start {
			labelStr := nameAndLabels[start+1 : end]
			labelPairs := strings.Split(labelStr, ",")
			
			for _, pair := range labelPairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
					labels[key] = val
				}
			}
		}
	}
	
	return &value, labels
}

func extractIBCMetrics(metrics []PrometheusMetric) gin.H {
	result := gin.H{
		"packet_flow": gin.H{},
		"stuck_packets": []gin.H{},
		"performance": gin.H{},
	}
	
	for _, metric := range metrics {
		switch metric.Name {
		case "ibc_effected_packets":
			result["packet_flow"].(gin.H)["effected"] = aggregateByChannel(metric.Values)
		case "ibc_uneffected_packets":
			result["packet_flow"].(gin.H)["uneffected"] = aggregateByChannel(metric.Values)
		case "ibc_stuck_packets":
			for _, v := range metric.Values {
				if v.Value > 0 {
					result["stuck_packets"] = append(result["stuck_packets"].([]gin.H), gin.H{
						"path": fmt.Sprintf("%s -> %s", v.Labels["src_chain"], v.Labels["dst_chain"]),
						"channel": v.Labels["src_channel"],
						"count": int(v.Value),
					})
				}
			}
		}
	}
	
	return result
}

func aggregateByChannel(values []PrometheusMetricValue) map[string]float64 {
	result := make(map[string]float64)
	
	for _, v := range values {
		key := fmt.Sprintf("%s/%s -> %s/%s", 
			v.Labels["src_channel"], 
			v.Labels["src_port"],
			v.Labels["dst_channel"],
			v.Labels["dst_port"])
		result[key] += v.Value
	}
	
	return result
}

func countMetricValues(metrics []PrometheusMetric, name string) int {
	count := 0
	for _, metric := range metrics {
		if metric.Name == name {
			for _, v := range metric.Values {
				count += int(v.Value)
			}
		}
	}
	return count
}