package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetRelayerStatus returns the status of both relayers
func (h *Handler) GetRelayerStatus(c *gin.Context) {
	status := gin.H{
		"hermes": h.checkHermesStatus(),
		"rly":    h.checkRlyStatus(),
	}

	c.JSON(http.StatusOK, status)
}

// StartHermes starts the Hermes relayer
func (h *Handler) StartHermes(c *gin.Context) {
	// Use supervisorctl to start Hermes
	cmd := exec.Command("supervisorctl", "start", "hermes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to start Hermes",
			"output": string(output),
		})
		return
	}

	// Also start Hermes REST API
	cmdRest := exec.Command("supervisorctl", "start", "hermes-rest")
	cmdRest.Run()

	// Broadcast status update
	h.broadcast <- gin.H{
		"type":    "relayer_status",
		"relayer": "hermes",
		"status":  "started",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Hermes started successfully",
		"output":  string(output),
	})
}

// StopHermes stops the Hermes relayer
func (h *Handler) StopHermes(c *gin.Context) {
	// Stop both Hermes and its REST API
	cmd := exec.Command("supervisorctl", "stop", "hermes", "hermes-rest")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to stop Hermes",
			"output": string(output),
		})
		return
	}

	// Broadcast status update
	h.broadcast <- gin.H{
		"type":    "relayer_status",
		"relayer": "hermes",
		"status":  "stopped",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Hermes stopped successfully",
		"output":  string(output),
	})
}

// StartGoRelayer starts the Go relayer
func (h *Handler) StartGoRelayer(c *gin.Context) {
	cmd := exec.Command("supervisorctl", "start", "go-relayer")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to start Go relayer",
			"output": string(output),
		})
		return
	}

	// Broadcast status update
	h.broadcast <- gin.H{
		"type":    "relayer_status",
		"relayer": "rly",
		"status":  "started",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Go relayer started successfully",
		"output":  string(output),
	})
}

// StopGoRelayer stops the Go relayer
func (h *Handler) StopGoRelayer(c *gin.Context) {
	cmd := exec.Command("supervisorctl", "stop", "go-relayer")
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to stop Go relayer",
			"output": string(output),
		})
		return
	}

	// Broadcast status update
	h.broadcast <- gin.H{
		"type":    "relayer_status",
		"relayer": "rly",
		"status":  "stopped",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Go relayer stopped successfully",
		"output":  string(output),
	})
}

// GetRelayerConfig returns the configuration for both relayers
func (h *Handler) GetRelayerConfig(c *gin.Context) {
	config := gin.H{}

	// Read Hermes config
	hermesConfigPath := os.Getenv("HERMES_CONFIG_PATH")
	if hermesConfigPath == "" {
		hermesConfigPath = "/home/relayer/.hermes/config.toml"
	}
	
	if hermesConfig, err := os.ReadFile(hermesConfigPath); err == nil {
		config["hermes"] = string(hermesConfig)
	} else {
		config["hermes"] = gin.H{"error": err.Error()}
	}

	// Read Go relayer config
	rlyConfigPath := os.Getenv("RLY_CONFIG_PATH")
	if rlyConfigPath == "" {
		rlyConfigPath = "/home/relayer/.relayer/config/config.yaml"
	}

	if rlyConfig, err := os.ReadFile(rlyConfigPath); err == nil {
		config["rly"] = string(rlyConfig)
	} else {
		config["rly"] = gin.H{"error": err.Error()}
	}

	c.JSON(http.StatusOK, config)
}

// UpdateRelayerConfig updates the configuration for a relayer
func (h *Handler) UpdateRelayerConfig(c *gin.Context) {
	var req struct {
		Relayer string `json:"relayer" binding:"required,oneof=hermes rly"`
		Config  string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var configPath string
	switch req.Relayer {
	case "hermes":
		configPath = os.Getenv("HERMES_CONFIG_PATH")
		if configPath == "" {
			configPath = "/home/relayer/.hermes/config.toml"
		}
	case "rly":
		configPath = os.Getenv("RLY_CONFIG_PATH")
		if configPath == "" {
			configPath = "/home/relayer/.relayer/config/config.yaml"
		}
	}

	// Backup existing config
	backupPath := fmt.Sprintf("%s.backup", configPath)
	if data, err := os.ReadFile(configPath); err == nil {
		os.WriteFile(backupPath, data, 0644)
	}

	// Write new config
	if err := os.WriteFile(configPath, []byte(req.Config), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("%s configuration updated successfully", req.Relayer),
		"backup":  backupPath,
	})
}

// Helper functions
func (h *Handler) checkHermesStatus() gin.H {
	status := gin.H{
		"running":     false,
		"rest_api":    false,
		"version":     "unknown",
		"config_path": os.Getenv("HERMES_CONFIG_PATH"),
	}

	// Check if process is running
	cmd := exec.Command("supervisorctl", "status", "hermes")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "RUNNING") {
		status["running"] = true
	}

	// Check REST API
	cmd = exec.Command("supervisorctl", "status", "hermes-rest")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "RUNNING") {
		status["rest_api"] = true

		// Get version from REST API
		if version, err := h.callHermesAPI("/version"); err == nil {
			status["version"] = version
		}
	}

	return status
}

func (h *Handler) checkRlyStatus() gin.H {
	status := gin.H{
		"running":     false,
		"version":     "unknown",
		"config_path": os.Getenv("RLY_CONFIG_PATH"),
	}

	// Check if process is running
	cmd := exec.Command("supervisorctl", "status", "go-relayer")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "RUNNING") {
		status["running"] = true
	}

	// Get version
	cmd = exec.Command("rly", "version")
	if output, err := cmd.Output(); err == nil {
		status["version"] = strings.TrimSpace(string(output))
	}

	return status
}