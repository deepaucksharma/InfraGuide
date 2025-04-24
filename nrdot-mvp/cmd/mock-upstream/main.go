package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Configuration for the mock-upstream service
type Config struct {
	HTTPPort               int    `json:"http_port"`
	MetricsPort            int    `json:"metrics_port"`
	LatencyMin             int    `json:"latency_min"`
	LatencyMax             int    `json:"latency_max"`
	ErrorRate              int    `json:"error_rate"`
	RateLimitErrorRate     int    `json:"rate_limit_error_rate"`
	SupportOutageSimulation bool   `json:"support_outage_simulation"`
	LogFile                string `json:"log_file"`
	LogLevel               string `json:"log_level"`
	VerboseLogging         bool   `json:"verbose_logging"`
}

// Stats tracks service statistics
type Stats struct {
	RequestsTotal     atomic.Int64
	RequestsFailed    atomic.Int64
	Outages           atomic.Int64
	OutageDuration    atomic.Int64
	BytesReceived     atomic.Int64
	ProcessingTimeNs  atomic.Int64
	LastRequestTimeNs atomic.Int64
}

// Global variables
var (
	config Config
	stats  Stats
	logger *log.Logger

	// Outage state
	inOutage       bool
	outageEndTime  time.Time
	outageLock     = make(chan struct{}, 1)
	outageComplete = make(chan struct{})

	// Prometheus metrics
	promRequestsTotal      *prometheus.CounterVec
	promRequestsFailed     *prometheus.CounterVec
	promBytesReceived      *prometheus.Counter
	promProcessingDuration *prometheus.HistogramVec
	promOutageStatus       *prometheus.Gauge
)

func main() {
	// Parse command line flags
	httpPort := flag.Int("port", 8080, "HTTP port for the main service")
	metricsPort := flag.Int("metrics-port", 8081, "HTTP port for Prometheus metrics")
	latencyMin := flag.Int("latency-min", 10, "Minimum artificial latency in ms")
	latencyMax := flag.Int("latency-max", 50, "Maximum artificial latency in ms")
	errorRate := flag.Int("error-rate", 0, "Rate of errors to return (0-100)")
	rateLimitErrorRate := flag.Int("rate-limit-errors", 0, "Rate of 429 errors to return (0-100)")
	supportOutage := flag.Bool("support-outage", true, "Whether to support outage simulation")
	logFile := flag.String("log-file", "", "Log file (empty for stdout)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize outageLock (buffered channel used as mutex)
	outageLock <- struct{}{}

	// Initialize config
	config = Config{
		HTTPPort:               *httpPort,
		MetricsPort:            *metricsPort,
		LatencyMin:             *latencyMin,
		LatencyMax:             *latencyMax,
		ErrorRate:              *errorRate,
		RateLimitErrorRate:     *rateLimitErrorRate,
		SupportOutageSimulation: *supportOutage,
		LogFile:                *logFile,
		LogLevel:               *logLevel,
		VerboseLogging:         *verbose,
	}

	// Check environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := fmt.Sscanf(port, "%d", &config.HTTPPort); err != nil {
			log.Printf("Invalid PORT environment variable: %s", port)
		}
	}
	if port := os.Getenv("METRICS_PORT"); port != "" {
		if p, err := fmt.Sscanf(port, "%d", &config.MetricsPort); err != nil {
			log.Printf("Invalid METRICS_PORT environment variable: %s", port)
		}
	}
	if errRate := os.Getenv("ERROR_RATE"); errRate != "" {
		if r, err := fmt.Sscanf(errRate, "%d", &config.ErrorRate); err != nil {
			log.Printf("Invalid ERROR_RATE environment variable: %s", errRate)
		}
	}
	if outage := os.Getenv("SUPPORT_OUTAGE_SIMULATION"); outage != "" {
		config.SupportOutageSimulation = (outage == "true" || outage == "1")
	}

	// Initialize logger
	if config.LogFile == "" {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer file.Close()
		logger = log.New(file, "", log.LstdFlags)
	}

	// Initialize Prometheus metrics
	initPrometheusMetrics()

	// Start metrics server
	go startMetricsServer()

	// Start HTTP server
	startHTTPServer()
}

func initPrometheusMetrics() {
	promRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_upstream_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"path", "method"},
	)

	promRequestsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mock_upstream_requests_failed_total",
			Help: "Total number of failed requests",
		},
		[]string{"path", "method", "reason"},
	)

	promBytesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mock_upstream_bytes_received_total",
			Help: "Total number of bytes received",
		},
	)

	promProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mock_upstream_processing_duration_seconds",
			Help:    "Time spent processing requests",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"path", "method"},
	)

	promOutageStatus = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mock_upstream_outage_status",
			Help: "Whether the service is in an outage state (0 = normal, 1 = outage)",
		},
	)

	// Register metrics
	prometheus.MustRegister(promRequestsTotal)
	prometheus.MustRegister(promRequestsFailed)
	prometheus.MustRegister(promBytesReceived)
	prometheus.MustRegister(promProcessingDuration)
	prometheus.MustRegister(promOutageStatus)
}

func startMetricsServer() {
	addr := fmt.Sprintf(":%d", config.MetricsPort)
	logger.Printf("Starting metrics server on %s", addr)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatalf("Failed to start metrics server: %v", err)
	}
}

func startHTTPServer() {
	// Create router
	mux := http.NewServeMux()

	// Add handlers
	mux.HandleFunc("/", handleRequest)
	mux.HandleFunc("/metrics", handleRequest)
	mux.HandleFunc("/traces", handleRequest)
	mux.HandleFunc("/logs", handleRequest)
	mux.HandleFunc("/profiles", handleRequest)
	mux.HandleFunc("/v1/metrics", handleRequest)
	mux.HandleFunc("/v1/traces", handleRequest)
	mux.HandleFunc("/v1/logs", handleRequest)
	mux.HandleFunc("/v1/profiles", handleRequest)
	mux.HandleFunc("/healthz", handleHealthCheck)
	mux.HandleFunc("/readyz", handleReadyCheck)
	
	// Outage control endpoint
	if config.SupportOutageSimulation {
		mux.HandleFunc("/outage", handleOutageControl)
	}

	// Start server
	addr := fmt.Sprintf(":%d", config.HTTPPort)
	logger.Printf("Starting mock upstream service on %s", addr)
	logger.Printf("Metrics available at :%d/metrics", config.MetricsPort)
	logger.Printf("Configuration: latency=%d-%dms, error-rate=%d%%, rate-limit-errors=%d%%",
		config.LatencyMin, config.LatencyMax, config.ErrorRate, config.RateLimitErrorRate)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Increment request counter
	stats.RequestsTotal.Add(1)
	promRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()

	// Check if we're in an outage
	if isInOutage() {
		// We're in an outage, return 503
		http.Error(w, "Service Unavailable: Simulated outage", http.StatusServiceUnavailable)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "outage").Inc()
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "read_error").Inc()
		return
	}

	// Update bytes received
	stats.BytesReceived.Add(int64(len(body)))
	promBytesReceived.Add(float64(len(body)))

	// Add artificial latency
	latency := config.LatencyMin
	if config.LatencyMax > config.LatencyMin {
		latency += rand.Intn(config.LatencyMax - config.LatencyMin)
	}
	time.Sleep(time.Duration(latency) * time.Millisecond)

	// Simulate errors based on error rate
	if config.ErrorRate > 0 && rand.Intn(100) < config.ErrorRate {
		http.Error(w, "Internal Server Error: Simulated error", http.StatusInternalServerError)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "error").Inc()
		return
	}

	// Simulate rate limiting errors
	if config.RateLimitErrorRate > 0 && rand.Intn(100) < config.RateLimitErrorRate {
		http.Error(w, "Too Many Requests: Rate limited", http.StatusTooManyRequests)
		promRequestsFailed.WithLabelValues(r.URL.Path, r.Method, "rate_limited").Inc()
		return
	}

	// Calculate processing time
	processingTime := time.Since(startTime)
	stats.ProcessingTimeNs.Add(processingTime.Nanoseconds())
	promProcessingDuration.WithLabelValues(r.URL.Path, r.Method).Observe(processingTime.Seconds())

	// Log request if verbose
	if config.VerboseLogging {
		logger.Printf("Processed request: %s %s %d bytes in %v",
			r.Method, r.URL.Path, len(body), processingTime)
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Health check is always healthy, even during outage (to distinguish from readiness)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func handleReadyCheck(w http.ResponseWriter, r *http.Request) {
	// Readiness check reflects outage state
	if isInOutage() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not ready","reason":"outage"}`))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	}
}

func handleOutageControl(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if outage simulation is supported
	if !config.SupportOutageSimulation {
		http.Error(w, "Outage simulation not supported", http.StatusServiceUnavailable)
		return
	}

	// Parse the request
	var req struct {
		Action          string `json:"action"`
		DurationSeconds int    `json:"duration_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process the action
	switch req.Action {
	case "start":
		// Start an outage
		durationSeconds := req.DurationSeconds
		if durationSeconds <= 0 {
			durationSeconds = 300 // Default to 5 minutes
		}

		if startOutage(durationSeconds) {
			// Outage started
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status":"outage started","duration_seconds":%d}`, durationSeconds)))
		} else {
			// Outage already in progress
			http.Error(w, "Outage already in progress", http.StatusConflict)
		}

	case "stop":
		// Stop the outage
		if stopOutage() {
			// Outage stopped
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"outage stopped"}`))
		} else {
			// No outage in progress
			http.Error(w, "No outage in progress", http.StatusConflict)
		}

	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

func startOutage(durationSeconds int) bool {
	// Try to acquire the outage lock
	select {
	case <-outageLock:
		// We got the lock
		if inOutage {
			// Already in outage, release lock and return false
			outageLock <- struct{}{}
			return false
		}

		// Start the outage
		inOutage = true
		outageEndTime = time.Now().Add(time.Duration(durationSeconds) * time.Second)
		promOutageStatus.Set(1)
		stats.Outages.Add(1)

		logger.Printf("Starting outage for %d seconds (until %s)",
			durationSeconds, outageEndTime.Format(time.RFC3339))

		// Release the lock
		outageLock <- struct{}{}

		// Start the auto-stop goroutine
		outageComplete = make(chan struct{})
		go func() {
			select {
			case <-time.After(time.Duration(durationSeconds) * time.Second):
				stopOutage()
			case <-outageComplete:
				// Outage manually stopped
				return
			}
		}()

		return true

	default:
		// Couldn't get the lock
		return false
	}
}

func stopOutage() bool {
	// Try to acquire the outage lock
	select {
	case <-outageLock:
		// We got the lock
		if !inOutage {
			// Not in outage, release lock and return false
			outageLock <- struct{}{}
			return false
		}

		// Stop the outage
		inOutage = false
		outageDuration := time.Since(outageEndTime.Add(-time.Duration(24) * time.Hour))
		stats.OutageDuration.Add(outageDuration.Milliseconds())
		promOutageStatus.Set(0)

		logger.Printf("Stopping outage (duration: %v)", outageDuration)

		// Signal the auto-stop goroutine to exit
		close(outageComplete)

		// Release the lock
		outageLock <- struct{}{}

		return true

	default:
		// Couldn't get the lock
		return false
	}
}

func isInOutage() bool {
	// Try to acquire the outage lock
	select {
	case <-outageLock:
		// We got the lock
		defer func() { outageLock <- struct{}{} }() // Release the lock when done

		if !inOutage {
			return false
		}

		// Check if the outage has expired
		if time.Now().After(outageEndTime) {
			// Outage has expired, stop it
			inOutage = false
			outageDuration := time.Since(outageEndTime.Add(-time.Duration(24) * time.Hour))
			stats.OutageDuration.Add(outageDuration.Milliseconds())
			promOutageStatus.Set(0)

			logger.Printf("Outage expired (duration: %v)", outageDuration)

			return false
		}

		return true

	default:
		// Couldn't get the lock, assume no outage
		return false
	}
}
