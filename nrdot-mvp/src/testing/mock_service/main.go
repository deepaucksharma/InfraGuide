package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Configuration for the mock service
type Config struct {
	// HTTP port to listen on
	Port int `json:"port"`
	
	// Prometheus metrics port
	MetricsPort int `json:"metrics_port"`
	
	// Artificial latency in milliseconds (min-max)
	LatencyMin int `json:"latency_min"`
	LatencyMax int `json:"latency_max"`
	
	// Error rate percentage (0-100)
	ErrorRate int `json:"error_rate"`
	
	// Whether to support the outage simulation mode
	SupportOutageSimulation bool `json:"support_outage_simulation"`
	
	// Whether to validate request data
	ValidateRequests bool `json:"validate_requests"`
	
	// Maximum request size in bytes
	MaxRequestSize int64 `json:"max_request_size"`
	
	// How many requests to process before responding
	SimultaneousRequests int `json:"simultaneous_requests"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:                  8080,
		MetricsPort:           8081,
		LatencyMin:            0,
		LatencyMax:            50,
		ErrorRate:             0,
		SupportOutageSimulation: true,
		ValidateRequests:      true,
		MaxRequestSize:        10 * 1024 * 1024, // 10 MiB
		SimultaneousRequests:  100,
	}
}

// Global variables
var (
	logger *zap.Logger
	config *Config
	
	// Runtime state
	inOutage       bool
	outageEndTime  time.Time
	requestsTotal  int64
	requestsFailed int64
	bytesTotal     int64
	
	// Request throttle for simulating max simultaneous requests
	requestSemaphore chan struct{}
	
	// Prometheus metrics
	promRequestsTotal   *prometheus.CounterVec
	promRequestsFailed  *prometheus.CounterVec
	promRequestLatency  *prometheus.HistogramVec
	promBytesReceived   *prometheus.Counter
	promOutageStatus    *prometheus.Gauge
	promCurrentRequests *prometheus.Gauge
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "Path to configuration file")
	port := flag.Int("port", 0, "HTTP port to listen on")
	metricsPort := flag.Int("metrics-port", 0, "Prometheus metrics port")
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
	if *port > 0 {
		config.Port = *port
	}
	if *metricsPort > 0 {
		config.MetricsPort = *metricsPort
	}
	
	// Override from environment
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := fmt.Sscanf(portStr, "%d", &config.Port); err != nil {
			logger.Warn("Invalid PORT environment variable", zap.Error(err))
		}
	}
	
	// Initialize request semaphore
	requestSemaphore = make(chan struct{}, config.SimultaneousRequests)
	
	// Initialize Prometheus metrics
	initPrometheusMetrics()
	
	// Start HTTP servers
	go startMetricsServer()
	go startHTTPServer()
	
	// Wait for shutdown signal
	waitForShutdown()
}

// loadConfig loads the configuration from a JSON file.
func loadConfig(path string, config *Config) error {
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

// initPrometheusMetrics initializes Prometheus metrics.
func initPrometheusMetrics() {
	promRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_service_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"path", "method"},
	)
	
	promRequestsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_service_requests_failed_total",
			Help: "Total number of failed requests",
		},
		[]string{"path", "method", "reason"},
	)
	
	promRequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mock_service_request_latency_ms",
			Help:    "Request latency in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		},
		[]string{"path", "method"},
	)
	
	promBytesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mock_service_bytes_received_total",
			Help: "Total number of bytes received",
		},
	)
	
	promOutageStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mock_service_outage_status",
			Help: "Whether the service is in an outage state (0 = normal, 1 = outage)",
		},
	)
	
	promCurrentRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mock_service_current_requests",
			Help: "Current number of active requests",
		},
	)
	
	// Register metrics
	prometheus.MustRegister(
		promRequestsTotal,
		promRequestsFailed,
		promRequestLatency,
		promBytesReceived,
		promOutageStatus,
		promCurrentRequests,
	)
}

// startMetricsServer starts the Prometheus metrics server.
func startMetricsServer() {
	addr := fmt.Sprintf(":%d", config.MetricsPort)
	logger.Info("Starting metrics server", zap.String("addr", addr))
	
	http.Handle("/metrics", promhttp.Handler())
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("Failed to start metrics server", zap.Error(err))
	}
}

// startHTTPServer starts the main HTTP server.
func startHTTPServer() {
	addr := fmt.Sprintf(":%d", config.Port)
	logger.Info("Starting HTTP server", zap.String("addr", addr))
	
	// Create router
	mux := http.NewServeMux()
	
	// Register handlers
	mux.HandleFunc("/v1/metrics", handleOTLP)
	mux.HandleFunc("/v1/traces", handleOTLP)
	mux.HandleFunc("/v1/logs", handleOTLP)
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/readyz", handleReadyCheck)
	mux.HandleFunc("/outage", handleOutageControl)
	
	// Start server
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

// handleOTLP handles OTLP requests.
func handleOTLP(w http.ResponseWriter, r *http.Request) {
	// Acquire semaphore
	select {
	case requestSemaphore <- struct{}{}:
		// Acquired semaphore
		defer func() {
			// Release semaphore
			<-requestSemaphore
		}()
	default:
		// Semaphore full, return service unavailable
		http.Error(w, "Service unavailable: too many requests", http.StatusServiceUnavailable)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "too_many_requests").Inc()
		return
	}
	
	// Update current requests gauge
	promCurrentRequests.Inc()
	defer promCurrentRequests.Dec()
	
	// Record request
	atomic.AddInt64(&requestsTotal, 1)
	promRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
	
	// Check if we're in an outage
	if isInOutage() {
		http.Error(w, "Service unavailable: simulated outage", http.StatusServiceUnavailable)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "outage").Inc()
		atomic.AddInt64(&requestsFailed, 1)
		return
	}
	
	// Check request size
	if config.MaxRequestSize > 0 && r.ContentLength > config.MaxRequestSize {
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "too_large").Inc()
		atomic.AddInt64(&requestsFailed, 1)
		return
	}
	
	// Start timing request
	startTime := time.Now()
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "read_error").Inc()
		atomic.AddInt64(&requestsFailed, 1)
		return
	}
	
	// Record bytes received
	bodySize := int64(len(body))
	atomic.AddInt64(&bytesTotal, bodySize)
	promBytesReceived.Add(float64(bodySize))
	
	// Validate request if enabled
	if config.ValidateRequests {
		if !validateOTLP(r.URL.Path, body) {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "invalid_format").Inc()
			atomic.AddInt64(&requestsFailed, 1)
			return
		}
	}
	
	// Add artificial latency
	if config.LatencyMax > 0 {
		latency := config.LatencyMin
		if config.LatencyMax > config.LatencyMin {
			latency += rand.Intn(config.LatencyMax - config.LatencyMin)
		}
		time.Sleep(time.Duration(latency) * time.Millisecond)
	}
	
	// Simulate error if configured
	if config.ErrorRate > 0 && rand.Intn(100) < config.ErrorRate {
		http.Error(w, "Simulated error", http.StatusInternalServerError)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "simulated_error").Inc()
		atomic.AddInt64(&requestsFailed, 1)
		return
	}
	
	// Calculate request latency
	latency := time.Since(startTime)
	promRequestLatency.WithLabelValues(r.URL.Path, r.Method).Observe(float64(latency.Milliseconds()))
	
	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"accepted":true}`))
}

// validateOTLP validates the format of OTLP requests.
func validateOTLP(path string, body []byte) bool {
	// Simple validation: check if body is valid JSON
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Debug("Invalid JSON in request", zap.Error(err))
		return false
	}
	
	// In a real implementation, we would validate the OTLP format more thoroughly
	return true
}

// handleHealthCheck handles health check requests.
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Always return healthy
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleReadyCheck handles readiness check requests.
func handleReadyCheck(w http.ResponseWriter, r *http.Request) {
	// Return not ready if in outage
	if isInOutage() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","reason":"outage"}`))
		return
	}
	
	// Otherwise return ready
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// handleOutageControl handles outage control requests.
func handleOutageControl(w http.ResponseWriter, r *http.Request) {
	// Check if outage simulation is supported
	if !config.SupportOutageSimulation {
		http.Error(w, "Outage simulation not supported", http.StatusBadRequest)
		return
	}
	
	// Check HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var req struct {
		Action   string `json:"action"`
		Duration int    `json:"duration_seconds"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Handle action
	switch req.Action {
	case "start":
		if req.Duration <= 0 {
			req.Duration = 60 // Default to 60 seconds
		}
		
		// Start outage
		startOutage(req.Duration)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"outage_started","duration_seconds":%d}`, req.Duration)))
		
	case "stop":
		// Stop outage
		stopOutage()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"outage_stopped"}`))
		
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// startOutage starts a simulated outage for the specified duration.
func startOutage(durationSeconds int) {
	inOutage = true
	outageEndTime = time.Now().Add(time.Duration(durationSeconds) * time.Second)
	promOutageStatus.Set(1)
	
	logger.Info("Started simulated outage",
		zap.Int("duration_seconds", durationSeconds),
		zap.Time("end_time", outageEndTime),
	)
	
	// Start a goroutine to automatically end the outage
	go func() {
		time.Sleep(time.Duration(durationSeconds) * time.Second)
		stopOutage()
	}()
}

// stopOutage stops the current simulated outage.
func stopOutage() {
	if !inOutage {
		return
	}
	
	inOutage = false
	promOutageStatus.Set(0)
	
	logger.Info("Stopped simulated outage")
}

// isInOutage checks if the service is currently in a simulated outage.
func isInOutage() bool {
	if !inOutage {
		return false
	}
	
	// Check if outage has expired
	if time.Now().After(outageEndTime) {
		stopOutage()
		return false
	}
	
	return true
}

// waitForShutdown waits for a shutdown signal.
func waitForShutdown() {
	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for signal
	sig := <-sigCh
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	
	// Give ongoing requests a chance to complete
	logger.Info("Waiting for ongoing requests to complete...")
	time.Sleep(1 * time.Second)
	
	logger.Info("Shutdown complete")
}
