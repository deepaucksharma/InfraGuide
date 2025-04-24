package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Configuration for the workload generator
type Config struct {
	// Target URL for sending data
	TargetURL string `json:"target_url"`
	
	// Number of concurrent workers
	Workers int `json:"workers"`
	
	// Rate limit (requests per second)
	RateLimit int `json:"rate_limit"`
	
	// Duration of the test in seconds
	Duration int `json:"duration"`
	
	// Send metrics
	SendMetrics bool `json:"send_metrics"`
	
	// Send traces
	SendTraces bool `json:"send_traces"`
	
	// Send logs
	SendLogs bool `json:"send_logs"`
	
	// Number of unique services to simulate
	UniqueServices int `json:"unique_services"`
	
	// Number of unique hosts to simulate
	UniqueHosts int `json:"unique_hosts"`
	
	// Number of unique instances to simulate
	UniqueInstances int `json:"unique_instances"`
	
	// Number of unique metrics to generate
	UniqueMetrics int `json:"unique_metrics"`
	
	// Number of unique traces to generate
	UniqueTraces int `json:"unique_traces"`
	
	// Number of unique logs to generate
	UniqueLogs int `json:"unique_logs"`
	
	// Number of dimensions per metric
	DimensionsPerMetric int `json:"dimensions_per_metric"`
	
	// Percentage of metrics that are critical priority (0-100)
	CriticalPercent int `json:"critical_percent"`
	
	// Percentage of metrics that are high priority (0-100)
	HighPercent int `json:"high_percent"`
	
	// Whether to introduce a random spike in cardinality
	CardinalitySpike bool `json:"cardinality_spike"`
	
	// If true, spike occurs at a random time. If false, occurs at SpikeTime
	RandomSpikeTime bool `json:"random_spike_time"`
	
	// Time in seconds when to introduce the spike
	SpikeTime int `json:"spike_time"`
	
	// Duration of the spike in seconds
	SpikeDuration int `json:"spike_duration"`
	
	// Factor to multiply cardinality during spike
	SpikeFactor int `json:"spike_factor"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		TargetURL:           "http://localhost:4318",
		Workers:             10,
		RateLimit:           1000,
		Duration:            300,
		SendMetrics:         true,
		SendTraces:          true,
		SendLogs:            true,
		UniqueServices:      10,
		UniqueHosts:         100,
		UniqueInstances:     1000,
		UniqueMetrics:       100,
		UniqueTraces:        50,
		UniqueLogs:          20,
		DimensionsPerMetric: 5,
		CriticalPercent:     10,
		HighPercent:         20,
		CardinalitySpike:    false,
		RandomSpikeTime:     true,
		SpikeTime:           60,
		SpikeDuration:       30,
		SpikeFactor:         10,
	}
}

// Constants
const (
	OTLPMetricsPath = "/v1/metrics"
	OTLPTracesPath  = "/v1/traces"
	OTLPLogsPath    = "/v1/logs"
)

// Global variables
var (
	logger *zap.Logger
	config *Config
	
	// Runtime state
	startTime      time.Time
	endTime        time.Time
	requestsSent   int64
	requestsFailed int64
	bytesTotal     int64
	latencyTotal   int64
	statsMutex     sync.Mutex
	
	// Workload state
	inSpike          bool
	spikeStartTime   time.Time
	spikeEndTime     time.Time
	normalDimensions int
	spikeDimensions  int
)

func main() {
	// Parse command line flags
	profileName := flag.String("profile", "default", "Name of the workload profile to use")
	targetURL := flag.String("target-url", "", "Target URL for the OTLP endpoint")
	workers := flag.Int("workers", 0, "Number of concurrent workers")
	duration := flag.Int("duration", 0, "Duration of the test in seconds")
	flag.Parse()
	
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	
	// Load configuration from profile
	config, err = loadProfile(*profileName)
	if err != nil {
		logger.Fatal("Failed to load profile", zap.Error(err))
	}
	
	// Override configuration with command line flags
	if *targetURL != "" {
		config.TargetURL = *targetURL
	}
	if *workers > 0 {
		config.Workers = *workers
	}
	if *duration > 0 {
		config.Duration = *duration
	}
	
	// Check if target URL is from environment variable
	if envURL := os.Getenv("TARGET_URL"); envURL != "" {
		config.TargetURL = envURL
	}
	
	// Initialize workload state
	startTime = time.Now()
	endTime = startTime.Add(time.Duration(config.Duration) * time.Second)
	
	// Set up cardinality spike if enabled
	if config.CardinalitySpike {
		normalDimensions = config.DimensionsPerMetric
		spikeDimensions = normalDimensions * config.SpikeFactor
		
		var spikeDelay time.Duration
		if config.RandomSpikeTime {
			spikeDelay = time.Duration(rand.Intn(config.Duration-config.SpikeDuration)) * time.Second
		} else {
			spikeDelay = time.Duration(config.SpikeTime) * time.Second
		}
		
		spikeStartTime = startTime.Add(spikeDelay)
		spikeEndTime = spikeStartTime.Add(time.Duration(config.SpikeDuration) * time.Second)
		
		logger.Info("Cardinality spike scheduled",
			zap.Time("startTime", spikeStartTime),
			zap.Time("endTime", spikeEndTime),
			zap.Int("normalDimensions", normalDimensions),
			zap.Int("spikeDimensions", spikeDimensions),
		)
	}
	
	// Log configuration
	logger.Info("Starting workload generator",
		zap.String("targetURL", config.TargetURL),
		zap.Int("workers", config.Workers),
		zap.Int("rateLimit", config.RateLimit),
		zap.Int("duration", config.Duration),
		zap.Time("startTime", startTime),
		zap.Time("endTime", endTime),
	)
	
	// Start stats reporter
	go statsReporter()
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < config.Workers; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	
	// Wait for completion
	wg.Wait()
	
	// Print final stats
	printStats(true)
	
	logger.Info("Workload generation completed")
}

// loadProfile loads a workload profile from a file.
func loadProfile(name string) (*Config, error) {
	// Default config
	config := DefaultConfig()
	
	// Try to load from file
	profilePath := fmt.Sprintf("profiles/%s.json", name)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		// If file not found, check if it's an environment variable
		logger.Warn("Profile file not found, using default with environment overrides",
			zap.String("profile", name),
			zap.Error(err),
		)
		return applyEnvironmentOverrides(config), nil
	}
	
	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}
	
	// Apply environment overrides
	return applyEnvironmentOverrides(config), nil
}

// applyEnvironmentOverrides applies environment variable overrides to the config.
func applyEnvironmentOverrides(config *Config) *Config {
	// Helper function to parse int from environment
	getEnvInt := func(key string, defaultVal int) int {
		if val, exists := os.LookupEnv(key); exists {
			if intVal, err := strconv.Atoi(val); err == nil {
				return intVal
			}
		}
		return defaultVal
	}
	
	// Helper function to parse bool from environment
	getEnvBool := func(key string, defaultVal bool) bool {
		if val, exists := os.LookupEnv(key); exists {
			return strings.ToLower(val) == "true" || val == "1"
		}
		return defaultVal
	}
	
	// Apply overrides
	if val, exists := os.LookupEnv("TARGET_URL"); exists {
		config.TargetURL = val
	}
	
	config.Workers = getEnvInt("WORKERS", config.Workers)
	config.RateLimit = getEnvInt("RATE_LIMIT", config.RateLimit)
	config.Duration = getEnvInt("DURATION", config.Duration)
	config.SendMetrics = getEnvBool("SEND_METRICS", config.SendMetrics)
	config.SendTraces = getEnvBool("SEND_TRACES", config.SendTraces)
	config.SendLogs = getEnvBool("SEND_LOGS", config.SendLogs)
	
	return config
}

// worker is a goroutine that generates and sends workload.
func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	
	logger.Info("Worker started", zap.Int("workerID", id))
	
	// Calculate interval between requests to achieve rate limit
	interval := time.Duration(1000000000 / (config.RateLimit / config.Workers)) * time.Nanosecond
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		// Check if test duration has elapsed
		if time.Now().After(endTime) {
			break
		}
		
		// Update spike status
		if config.CardinalitySpike {
			now := time.Now()
			if now.After(spikeStartTime) && now.Before(spikeEndTime) && !inSpike {
				inSpike = true
				logger.Info("Entering cardinality spike",
					zap.Time("time", now),
					zap.Int("dimensions", spikeDimensions),
				)
			} else if now.After(spikeEndTime) && inSpike {
				inSpike = false
				logger.Info("Exiting cardinality spike",
					zap.Time("time", now),
					zap.Int("dimensions", normalDimensions),
				)
			}
		}
		
		// Send telemetry data
		sendData()
	}
	
	logger.Info("Worker finished", zap.Int("workerID", id))
}

// sendData generates and sends telemetry data.
func sendData() {
	// Determine what to send based on configuration and random selection
	sendTypes := make([]string, 0, 3)
	if config.SendMetrics {
		sendTypes = append(sendTypes, "metrics")
	}
	if config.SendTraces {
		sendTypes = append(sendTypes, "traces")
	}
	if config.SendLogs {
		sendTypes = append(sendTypes, "logs")
	}
	
	if len(sendTypes) == 0 {
		return
	}
	
	// Randomly select one type to send
	dataType := sendTypes[rand.Intn(len(sendTypes))]
	
	switch dataType {
	case "metrics":
		sendMetrics()
	case "traces":
		sendTraces()
	case "logs":
		sendLogs()
	}
}

// sendMetrics generates and sends metrics data.
func sendMetrics() {
	// Generate metrics data
	payload := generateMetricsPayload()
	
	// Send to OTLP endpoint
	sendOTLP(OTLPMetricsPath, payload)
}

// sendTraces generates and sends traces data.
func sendTraces() {
	// Generate traces data
	payload := generateTracesPayload()
	
	// Send to OTLP endpoint
	sendOTLP(OTLPTracesPath, payload)
}

// sendLogs generates and sends logs data.
func sendLogs() {
	// Generate logs data
	payload := generateLogsPayload()
	
	// Send to OTLP endpoint
	sendOTLP(OTLPLogsPath, payload)
}

// sendOTLP sends data to the OTLP endpoint.
func sendOTLP(path string, payload []byte) {
	url := config.TargetURL + path
	
	// Record request time
	startTime := time.Now()
	
	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error("Failed to create request", zap.Error(err))
		recordFailure()
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Determine priority level
	priorityLevel := determinePriority()
	if priorityLevel != "" {
		req.Header.Set("X-Priority", priorityLevel)
	}
	
	// Send request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	
	// Calculate latency
	latency := time.Since(startTime)
	
	// Handle errors
	if err != nil {
		logger.Error("Request failed",
			zap.Error(err),
			zap.String("url", url),
			zap.Duration("latency", latency),
		)
		recordFailure()
		return
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		logger.Error("Request failed",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("url", url),
			zap.Duration("latency", latency),
		)
		recordFailure()
		return
	}
	
	// Record success
	recordSuccess(len(payload), latency)
}

// determinePriority randomly assigns a priority level based on configuration.
func determinePriority() string {
	roll := rand.Intn(100)
	
	if roll < config.CriticalPercent {
		return "critical"
	} else if roll < config.CriticalPercent+config.HighPercent {
		return "high"
	}
	
	return "normal"
}

// generateMetricsPayload generates a metrics payload.
func generateMetricsPayload() []byte {
	// In a real implementation, this would generate actual OTLP metrics
	// For simplicity, we'll just return a placeholder
	dimensions := config.DimensionsPerMetric
	if inSpike {
		dimensions = spikeDimensions
	}
	
	// Generate a payload with the specified dimensions
	// This is a simplified placeholder
	payload := fmt.Sprintf(`{
		"resourceMetrics": [
			{
				"resource": {
					"attributes": [
						{"key": "service.name", "value": {"stringValue": "service-%d"}},
						{"key": "host.name", "value": {"stringValue": "host-%d"}}
					]
				},
				"scopeMetrics": [
					{
						"metrics": [
							{
								"name": "metric-%d",
								"gauge": {
									"dataPoints": [
										{
											"timeUnixNano": "%d",
											"asDouble": %f,
											"attributes": [
												%s
											]
										}
									]
								}
							}
						]
					}
				]
			}
		]
	}`,
		rand.Intn(config.UniqueServices),
		rand.Intn(config.UniqueHosts),
		rand.Intn(config.UniqueMetrics),
		time.Now().UnixNano(),
		rand.Float64()*100,
		generateAttributes(dimensions),
	)
	
	return []byte(payload)
}

// generateAttributes generates random attributes for metrics.
func generateAttributes(count int) string {
	attrs := make([]string, count)
	
	for i := 0; i < count; i++ {
		attrs[i] = fmt.Sprintf(`{"key": "dim%d", "value": {"stringValue": "val-%d"}}`, 
			i, rand.Intn(1000))
	}
	
	return strings.Join(attrs, ",")
}

// generateTracesPayload generates a traces payload.
func generateTracesPayload() []byte {
	// In a real implementation, this would generate actual OTLP traces
	// For simplicity, we'll just return a placeholder
	return []byte(`{"resourceSpans":[]}`)
}

// generateLogsPayload generates a logs payload.
func generateLogsPayload() []byte {
	// In a real implementation, this would generate actual OTLP logs
	// For simplicity, we'll just return a placeholder
	return []byte(`{"resourceLogs":[]}`)
}

// recordSuccess records a successful request.
func recordSuccess(bytes int, latency time.Duration) {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	requestsSent++
	bytesTotal += int64(bytes)
	latencyTotal += latency.Microseconds()
}

// recordFailure records a failed request.
func recordFailure() {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	requestsFailed++
}

// statsReporter periodically reports statistics.
func statsReporter() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if time.Now().After(endTime) {
			return
		}
		
		printStats(false)
	}
}

// printStats prints current statistics.
func printStats(final bool) {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	elapsed := time.Since(startTime)
	rps := float64(requestsSent) / elapsed.Seconds()
	
	var avgLatency float64
	if requestsSent > 0 {
		avgLatency = float64(latencyTotal) / float64(requestsSent)
	}
	
	status := "progress"
	if final {
		status = "final"
	}
	
	logger.Info(fmt.Sprintf("Workload stats (%s)", status),
		zap.Duration("elapsed", elapsed),
		zap.Int64("requestsSent", requestsSent),
		zap.Int64("requestsFailed", requestsFailed),
		zap.Float64("rps", rps),
		zap.Float64("avgLatencyMs", avgLatency/1000),
		zap.Int64("bytesTotal", bytesTotal),
		zap.Bool("inCardinalitySpike", inSpike),
	)
}
