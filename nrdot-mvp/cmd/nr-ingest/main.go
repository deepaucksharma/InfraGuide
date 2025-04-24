package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Configuration for the nr-ingest mock service
type Config struct {
	HTTPPort       int    `json:"http_port"`
	MetricsPort    int    `json:"metrics_port"`
	LogFile        string `json:"log_file"`
	LogLevel       string `json:"log_level"`
	VerboseLogging bool   `json:"verbose_logging"`
}

// Stats tracks ingest statistics
type Stats struct {
	MetricsReceived   atomic.Int64
	TracesReceived    atomic.Int64
	LogsReceived      atomic.Int64
	ProfilesReceived  atomic.Int64
	BytesReceived     atomic.Int64
	TotalRequests     atomic.Int64
	FailedRequests    atomic.Int64
	ProcessingTimeNs  atomic.Int64
	LastRequestTimeNs atomic.Int64
}

// Global variables
var (
	config Config
	stats  Stats
	logger *log.Logger

	// Prometheus metrics
	promRequestsTotal      *prometheus.CounterVec
	promBytesReceived      *prometheus.Counter
	promProcessingDuration *prometheus.HistogramVec
	promTelemetryItems     *prometheus.CounterVec
)

func main() {
	// Parse command line flags
	httpPort := flag.Int("port", 4317, "HTTP port for the OTLP endpoint")
	metricsPort := flag.Int("metrics-port", 8889, "HTTP port for Prometheus metrics")
	logFile := flag.String("log-file", "", "Log file (empty for stdout)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize config
	config = Config{
		HTTPPort:       *httpPort,
		MetricsPort:    *metricsPort,
		LogFile:        *logFile,
		LogLevel:       *logLevel,
		VerboseLogging: *verbose,
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
			Name: "nr_ingest_requests_total",
			Help: "Total number of requests received by signal type",
		},
		[]string{"type"},
	)

	promBytesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "nr_ingest_bytes_received_total",
			Help: "Total number of bytes received",
		},
	)

	promProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nr_ingest_processing_duration_seconds",
			Help:    "Time spent processing requests",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"type"},
	)

	promTelemetryItems = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nr_ingest_telemetry_items_total",
			Help: "Total number of telemetry items received",
		},
		[]string{"type"},
	)

	// Register metrics
	prometheus.MustRegister(promRequestsTotal)
	prometheus.MustRegister(promBytesReceived)
	prometheus.MustRegister(promProcessingDuration)
	prometheus.MustRegister(promTelemetryItems)
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

	// OTLP routes
	mux.HandleFunc("/v1/metrics", handleOTLPRequest("metrics"))
	mux.HandleFunc("/v1/traces", handleOTLPRequest("traces"))
	mux.HandleFunc("/v1/logs", handleOTLPRequest("logs"))
	mux.HandleFunc("/v1/profiles", handleOTLPRequest("profiles"))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Start server
	addr := fmt.Sprintf(":%d", config.HTTPPort)
	logger.Printf("Starting NR Ingest mock server on %s", addr)
	logger.Printf("Metrics available at :%d/metrics", config.MetricsPort)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func handleOTLPRequest(signalType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Verify method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			stats.FailedRequests.Add(1)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			stats.FailedRequests.Add(1)
			return
		}

		// Decompress if needed (in a real implementation)
		// For now, we'll just count the raw bytes
		bodySize := int64(len(body))
		stats.BytesReceived.Add(bodySize)
		promBytesReceived.Add(float64(bodySize))

		// Process based on signal type
		switch signalType {
		case "metrics":
			stats.MetricsReceived.Add(1)
			// Parse metrics (simplified for mock)
			countMetrics(body)
		case "traces":
			stats.TracesReceived.Add(1)
			// Parse traces (simplified for mock)
			countTraces(body)
		case "logs":
			stats.LogsReceived.Add(1)
			// Parse logs (simplified for mock)
			countLogs(body)
		case "profiles":
			stats.ProfilesReceived.Add(1)
			// Parse profiles (simplified for mock)
			countProfiles(body)
		}

		// Update stats
		stats.TotalRequests.Add(1)
		promRequestsTotal.WithLabelValues(signalType).Inc()

		// Calculate processing time
		processingTime := time.Since(startTime)
		stats.ProcessingTimeNs.Add(processingTime.Nanoseconds())
		stats.LastRequestTimeNs.Store(time.Now().UnixNano())
		promProcessingDuration.WithLabelValues(signalType).Observe(processingTime.Seconds())

		// Log request if verbose
		if config.VerboseLogging {
			logger.Printf("Received %s request: %d bytes, processed in %v", 
				signalType, bodySize, processingTime)
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}
}

// Parse and count metrics (simplified implementation)
func countMetrics(body []byte) {
	// In a real implementation, parse OTLP metrics protobuf
	// For this mock, we'll just count as 1 batch
	promTelemetryItems.WithLabelValues("metrics").Inc()
	
	// Log request data for debugging
	if config.VerboseLogging {
		logger.Printf("Processed metrics batch")
	}
}

// Parse and count traces (simplified implementation)
func countTraces(body []byte) {
	// In a real implementation, parse OTLP traces protobuf
	// For this mock, we'll just count as 1 batch
	promTelemetryItems.WithLabelValues("traces").Inc()
	
	// Log request data for debugging
	if config.VerboseLogging {
		logger.Printf("Processed traces batch")
	}
}

// Parse and count logs (simplified implementation)
func countLogs(body []byte) {
	// In a real implementation, parse OTLP logs protobuf
	// For this mock, we'll just count as 1 batch
	promTelemetryItems.WithLabelValues("logs").Inc()
	
	// Log request data for debugging
	if config.VerboseLogging {
		logger.Printf("Processed logs batch")
	}
}

// Parse and count profiles (simplified implementation)
func countProfiles(body []byte) {
	// In a real implementation, parse OTLP profiles protobuf
	// For this mock, we'll just count as 1 batch
	promTelemetryItems.WithLabelValues("profiles").Inc()
	
	// Log request data for debugging
	if config.VerboseLogging {
		logger.Printf("Processed profiles batch")
	}
}
