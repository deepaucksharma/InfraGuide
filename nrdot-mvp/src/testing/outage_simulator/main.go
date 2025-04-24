package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OutageConfig defines the configuration for the outage simulation.
type OutageConfig struct {
	// Target service to simulate outage for
	TargetService string `json:"target_service"`
	
	// Target URL for the outage control endpoint
	TargetURL string `json:"target_url"`
	
	// Duration of the outage in seconds
	OutageDuration int `json:"outage_duration"`
	
	// Type of outage to simulate
	OutageType string `json:"outage_type"`
	
	// Whether to wait for the outage to complete before exiting
	WaitForCompletion bool `json:"wait_for_completion"`
	
	// Whether to verify DLQ functionality after the outage
	VerifyDLQ bool `json:"verify_dlq"`
	
	// Location of the DLQ files to verify
	DLQDirectory string `json:"dlq_directory"`
	
	// Docker container to target (if using container_stop outage type)
	DockerContainer string `json:"docker_container"`
	
	// Whether to restart the container automatically after outage
	AutoRestart bool `json:"auto_restart"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *OutageConfig {
	return &OutageConfig{
		TargetService:     "mock-service",
		TargetURL:         "http://localhost:8080/outage",
		OutageDuration:    60,
		OutageType:        "api",
		WaitForCompletion: true,
		VerifyDLQ:         true,
		DLQDirectory:      "/var/lib/otel/dlq",
		DockerContainer:   "nrdot-mvp_mock-service_1",
		AutoRestart:       true,
	}
}

// Global variables
var (
	logger *zap.Logger
	config *OutageConfig
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	targetService := flag.String("target", "", "Target service to simulate outage for")
	outageType := flag.String("type", "", "Type of outage to simulate (api, container_stop, network)")
	duration := flag.Int("duration", 0, "Duration of the outage in seconds")
	targetURL := flag.String("url", "", "Target URL for outage control")
	flag.Parse()
	
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	
	// Load configuration
	config = DefaultConfig()
	if *configFile != "" {
		if err := loadConfig(*configFile, config); err != nil {
			logger.Fatal("Failed to load configuration", zap.Error(err))
		}
	}
	
	// Override with command-line flags
	if *targetService != "" {
		config.TargetService = *targetService
	}
	if *outageType != "" {
		config.OutageType = *outageType
	}
	if *duration > 0 {
		config.OutageDuration = *duration
	}
	if *targetURL != "" {
		config.TargetURL = *targetURL
	}
	
	// Override from environment
	if envTarget := os.Getenv("TARGET_SERVICE"); envTarget != "" {
		config.TargetService = envTarget
	}
	if envType := os.Getenv("OUTAGE_TYPE"); envType != "" {
		config.OutageType = envType
	}
	if envURL := os.Getenv("TARGET_URL"); envURL != "" {
		config.TargetURL = envURL
	}
	
	// Log configuration
	logger.Info("Starting outage simulation",
		zap.String("targetService", config.TargetService),
		zap.String("outageType", config.OutageType),
		zap.Int("duration", config.OutageDuration),
		zap.String("targetURL", config.TargetURL),
	)
	
	// Simulate outage
	if err := simulateOutage(); err != nil {
		logger.Fatal("Failed to simulate outage", zap.Error(err))
	}
	
	// Wait for completion if configured
	if config.WaitForCompletion {
		logger.Info("Waiting for outage to complete...",
			zap.Int("durationSeconds", config.OutageDuration),
		)
		time.Sleep(time.Duration(config.OutageDuration) * time.Second)
		logger.Info("Outage completed")
	}
	
	// Verify DLQ if configured
	if config.VerifyDLQ {
		if err := verifyDLQ(); err != nil {
			logger.Error("DLQ verification failed", zap.Error(err))
		} else {
			logger.Info("DLQ verification successful")
		}
	}
}

// loadConfig loads the configuration from a JSON file.
func loadConfig(path string, config *OutageConfig) error {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// simulateOutage simulates an outage based on the configuration.
func simulateOutage() error {
	switch config.OutageType {
	case "api":
		return simulateAPIOutage()
	case "container_stop":
		return simulateContainerStopOutage()
	case "network":
		return simulateNetworkOutage()
	default:
		return fmt.Errorf("unsupported outage type: %s", config.OutageType)
	}
}

// simulateAPIOutage simulates an outage using the API endpoint.
func simulateAPIOutage() error {
	// Create request payload
	payload := map[string]interface{}{
		"action":          "start",
		"duration_seconds": config.OutageDuration,
	}
	
	// Convert to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Send request
	resp, err := http.Post(config.TargetURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send outage request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("outage request failed with status: %d", resp.StatusCode)
	}
	
	logger.Info("API outage started", 
		zap.Int("duration", config.OutageDuration),
		zap.String("targetURL", config.TargetURL),
	)
	
	return nil
}

// simulateContainerStopOutage simulates an outage by stopping a Docker container.
func simulateContainerStopOutage() error {
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker command not found: %w", err)
	}
	
	// Stop the container
	stopCmd := exec.Command("docker", "stop", config.DockerContainer)
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	
	logger.Info("Container stopped", 
		zap.String("container", config.DockerContainer),
		zap.Int("duration", config.OutageDuration),
	)
	
	// If auto-restart is enabled, schedule the restart
	if config.AutoRestart {
		go func() {
			// Wait for outage duration
			time.Sleep(time.Duration(config.OutageDuration) * time.Second)
			
			// Restart the container
			startCmd := exec.Command("docker", "start", config.DockerContainer)
			if err := startCmd.Run(); err != nil {
				logger.Error("Failed to restart container", 
					zap.String("container", config.DockerContainer),
					zap.Error(err),
				)
				return
			}
			
			logger.Info("Container restarted", 
				zap.String("container", config.DockerContainer),
			)
		}()
	}
	
	return nil
}

// simulateNetworkOutage simulates a network outage using iptables (Linux only).
func simulateNetworkOutage() error {
	// Check if iptables is available
	if _, err := exec.LookPath("iptables"); err != nil {
		return fmt.Errorf("iptables command not found (requires Linux): %w", err)
	}
	
	// Parse target service to extract host and port
	parts := strings.Split(config.TargetService, ":")
	host := parts[0]
	port := "80"
	if len(parts) > 1 {
		port = parts[1]
	}
	
	// Add iptables rule to block traffic
	blockCmd := exec.Command("iptables", "-A", "OUTPUT", "-d", host, "-p", "tcp", "--dport", port, "-j", "DROP")
	if err := blockCmd.Run(); err != nil {
		return fmt.Errorf("failed to add iptables rule: %w", err)
	}
	
	logger.Info("Network outage started", 
		zap.String("host", host),
		zap.String("port", port),
		zap.Int("duration", config.OutageDuration),
	)
	
	// Schedule rule removal
	go func() {
		// Wait for outage duration
		time.Sleep(time.Duration(config.OutageDuration) * time.Second)
		
		// Remove iptables rule
		unblockCmd := exec.Command("iptables", "-D", "OUTPUT", "-d", host, "-p", "tcp", "--dport", port, "-j", "DROP")
		if err := unblockCmd.Run(); err != nil {
			logger.Error("Failed to remove iptables rule", 
				zap.String("host", host),
				zap.String("port", port),
				zap.Error(err),
			)
			return
		}
		
		logger.Info("Network outage ended", 
			zap.String("host", host),
			zap.String("port", port),
		)
	}()
	
	return nil
}

// verifyDLQ verifies that data was properly saved to the DLQ during the outage.
func verifyDLQ() error {
	// In a real implementation, this would check that data was properly written to the DLQ
	// during the outage and verify the integrity using SHA-256
	
	// This is a placeholder implementation
	logger.Info("Verifying DLQ", zap.String("directory", config.DLQDirectory))
	
	// Check if DLQ directory exists
	info, err := os.Stat(config.DLQDirectory)
	if err != nil {
		return fmt.Errorf("failed to access DLQ directory: %w", err)
	}
	
	if !info.IsDir() {
		return fmt.Errorf("DLQ path is not a directory: %s", config.DLQDirectory)
	}
	
	// List files in DLQ directory
	files, err := os.ReadDir(config.DLQDirectory)
	if err != nil {
		return fmt.Errorf("failed to read DLQ directory: %w", err)
	}
	
	// Check if there are any DLQ files
	var dlqFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dlq") {
			dlqFiles = append(dlqFiles, file.Name())
		}
	}
	
	if len(dlqFiles) == 0 {
		return fmt.Errorf("no DLQ files found in directory: %s", config.DLQDirectory)
	}
	
	logger.Info("Found DLQ files", 
		zap.Int("count", len(dlqFiles)),
		zap.Strings("files", dlqFiles),
	)
	
	// In a full implementation, we would:
	// 1. Read each DLQ file
	// 2. Verify the SHA-256 signatures
	// 3. Check timestamps to ensure data was written during the outage
	// 4. Verify the content format
	
	return nil
}
