package adaptivedegradationmanager

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
)

// processor implements the AdaptiveDegradationManager processor.
type processor struct {
	logger            *zap.Logger
	config            *Config
	metricsConsumer   consumer.Metrics
	tracesConsumer    consumer.Traces
	logsConsumer      consumer.Logs
	
	// State
	currentLevel      *atomic.Int32
	lastLevelChange   time.Time
	stateMutex        sync.RWMutex
	
	// Metrics
	memoryUtilization float64
	queueUtilization  float64
	cpuUtilization    float64
	errorRate         float64
	latencyP99        float64
	
	// Action state
	sampleRate        float64
	batchMultiplier   int
	scrapeMultiplier  int
	dropDebug         bool
	dropMetrics       bool
	
	// Prometheus metrics
	levelGauge        prometheus.Gauge
	actionsCounter    *prometheus.CounterVec
	droppedCounter    *prometheus.CounterVec
	stateGauge        *prometheus.GaugeVec
	
	// Metrics poller
	cancelPoller      context.CancelFunc
}

// newProcessor creates a new AdaptiveDegradationManager processor.
func newProcessor(
	logger *zap.Logger,
	config *Config,
	nextConsumer interface{},
) (*processor, error) {
	p := &processor{
		logger:          logger,
		config:          config,
		currentLevel:    atomic.NewInt32(0),
		lastLevelChange: time.Now(),
		sampleRate:      1.0,
		batchMultiplier: 1,
		scrapeMultiplier: 1,
		dropDebug:       false,
		dropMetrics:     false,
	}
	
	// Set the appropriate consumer based on the type
	switch c := nextConsumer.(type) {
	case consumer.Metrics:
		p.metricsConsumer = c
	case consumer.Traces:
		p.tracesConsumer = c
	case consumer.Logs:
		p.logsConsumer = c
	}
	
	// Initialize Prometheus metrics
	p.initMetrics()
	
	return p, nil
}

// initMetrics initializes Prometheus metrics.
func (p *processor) initMetrics() {
	p.levelGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "otelcol_adm_current_level",
		Help: "Current adaptive degradation level (0 = normal, higher = more degraded)",
	})
	
	p.actionsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otelcol_adm_actions_total",
			Help: "Count of adaptive degradation actions taken",
		},
		[]string{"action"},
	)
	
	p.droppedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otelcol_adm_dropped_total",
			Help: "Count of items dropped due to adaptive degradation",
		},
		[]string{"telemetry_type"},
	)
	
	p.stateGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "otelcol_adm_state",
			Help: "Current state values monitored by adaptive degradation manager",
		},
		[]string{"metric"},
	)
	
	// Register metrics
	registry := prometheus.DefaultRegisterer
	registry.MustRegister(p.levelGauge)
	registry.MustRegister(p.actionsCounter)
	registry.MustRegister(p.droppedCounter)
	registry.MustRegister(p.stateGauge)
}

// Start starts the processor, including metrics collection.
func (p *processor) Start(ctx context.Context, host component.Host) error {
	ctx, cancel := context.WithCancel(ctx)
	p.cancelPoller = cancel
	
	// Start a goroutine to poll metrics and update degradation level
	go p.pollMetrics(ctx)
	
	return nil
}

// Shutdown stops the processor.
func (p *processor) Shutdown(ctx context.Context) error {
	if p.cancelPoller != nil {
		p.cancelPoller()
	}
	return nil
}

// pollMetrics periodically polls metrics and updates the degradation level.
func (p *processor) pollMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.CheckInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
			p.assessDegradationLevel()
		}
	}
}

// updateMetrics updates the current metrics.
func (p *processor) updateMetrics() {
	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	totalMemory := float64(memStats.Sys)
	usedMemory := float64(memStats.HeapInuse + memStats.StackInuse)
	p.memoryUtilization = (usedMemory / totalMemory) * 100
	
	// Update metrics gauges
	p.stateGauge.WithLabelValues("memory_utilization").Set(p.memoryUtilization)
	p.stateGauge.WithLabelValues("queue_utilization").Set(p.queueUtilization)
	p.stateGauge.WithLabelValues("cpu_utilization").Set(p.cpuUtilization)
	p.stateGauge.WithLabelValues("error_rate").Set(p.errorRate)
	p.stateGauge.WithLabelValues("latency_p99").Set(p.latencyP99)
}

// assessDegradationLevel determines the appropriate degradation level based on current metrics.
func (p *processor) assessDegradationLevel() {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()
	
	currentLevel := int(p.currentLevel.Load())
	newLevel := 0
	
	// Check triggers to determine the appropriate level
	if p.memoryUtilization >= float64(p.config.Triggers.MemoryUtilizationHigh) ||
	   p.queueUtilization >= float64(p.config.Triggers.QueueUtilizationHigh) ||
	   p.cpuUtilization >= float64(p.config.Triggers.CPUUtilizationHigh) ||
	   p.errorRate >= float64(p.config.Triggers.ErrorRateHigh) ||
	   p.latencyP99 >= float64(p.config.Triggers.LatencyP99High) {
		
		// Determine the appropriate level based on severity
		if p.memoryUtilization >= 90 || p.queueUtilization >= 90 {
			newLevel = 3 // Most severe
		} else if p.memoryUtilization >= 80 || p.queueUtilization >= 80 {
			newLevel = 2
		} else {
			newLevel = 1
		}
	}
	
	// Only decrease level if cooldown period has passed
	if newLevel < currentLevel && time.Since(p.lastLevelChange) < time.Duration(p.config.CooldownPeriod)*time.Second {
		return
	}
	
	// Update level if changed
	if newLevel != currentLevel {
		p.setDegradationLevel(newLevel)
	}
}

// setDegradationLevel sets a new degradation level and applies the associated actions.
func (p *processor) setDegradationLevel(level int) {
	oldLevel := int(p.currentLevel.Load())
	p.currentLevel.Store(int32(level))
	p.lastLevelChange = time.Now()
	p.levelGauge.Set(float64(level))
	
	p.logger.Info("Changing adaptive degradation level",
		zap.Int("old_level", oldLevel),
		zap.Int("new_level", level),
		zap.Float64("memory_utilization", p.memoryUtilization),
		zap.Float64("queue_utilization", p.queueUtilization))
	
	// Reset all actions
	p.sampleRate = 1.0
	p.batchMultiplier = 1
	p.scrapeMultiplier = 1
	p.dropDebug = false
	p.dropMetrics = false
	
	// Apply actions for the new level
	if level > 0 && level <= len(p.config.Levels) {
		levelIdx := level - 1
		for _, action := range p.config.Levels[levelIdx].Actions {
			p.applyAction(action)
			p.actionsCounter.WithLabelValues(action).Inc()
		}
	}
}

// applyAction applies a specific degradation action.
func (p *processor) applyAction(action string) {
	switch action {
	case "inc_batch":
		p.batchMultiplier = 2
	case "stretch_scrape":
		p.scrapeMultiplier = 2
	case "enable_sampling":
		p.sampleRate = 0.5
	case "drop_debug":
		p.dropDebug = true
	case "drop_metrics":
		p.dropMetrics = true
	}
}

// ConsumeMetrics implements the metrics consumer interface.
func (p *processor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	level := int(p.currentLevel.Load())
	
	// Apply degradation if level > 0
	if level > 0 {
		if p.dropMetrics {
			p.droppedCounter.WithLabelValues("metrics").Inc()
			return nil
		}
		
		// Apply sampling if enabled
		if p.sampleRate < 1.0 && rand.Float64() > p.sampleRate {
			p.droppedCounter.WithLabelValues("metrics").Inc()
			return nil
		}
	}
	
	return p.metricsConsumer.ConsumeMetrics(ctx, md)
}

// ConsumeTraces implements the traces consumer interface.
func (p *processor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	level := int(p.currentLevel.Load())
	
	// Apply degradation if level > 0
	if level > 0 {
		// Apply sampling if enabled
		if p.sampleRate < 1.0 && rand.Float64() > p.sampleRate {
			p.droppedCounter.WithLabelValues("traces").Inc()
			return nil
		}
		
		// Filter debug spans if dropDebug is enabled
		if p.dropDebug {
			td = filterDebugSpans(td)
		}
	}
	
	return p.tracesConsumer.ConsumeTraces(ctx, td)
}

// ConsumeLogs implements the logs consumer interface.
func (p *processor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	level := int(p.currentLevel.Load())
	
	// Apply degradation if level > 0
	if level > 0 {
		// Apply sampling if enabled
		if p.sampleRate < 1.0 && rand.Float64() > p.sampleRate {
			p.droppedCounter.WithLabelValues("logs").Inc()
			return nil
		}
		
		// Filter debug logs if dropDebug is enabled
		if p.dropDebug {
			ld = filterDebugLogs(ld)
		}
	}
	
	return p.logsConsumer.ConsumeLogs(ctx, ld)
}

// filterDebugSpans removes spans with debug flag or low severity.
func filterDebugSpans(td ptrace.Traces) ptrace.Traces {
	// In a real implementation, this would check for debug-level spans
	// and remove them from the traces data.
	// For simplicity, we'll just return the original traces.
	return td
}

// filterDebugLogs removes debug-level logs.
func filterDebugLogs(ld plog.Logs) plog.Logs {
	// Create a new logs collection
	filtered := plog.NewLogs()
	
	// Iterate through resource logs
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		resourceLogs := ld.ResourceLogs().At(i)
		
		// Create a new resource logs entry
		newResourceLogs := filtered.ResourceLogs().AppendEmpty()
		resourceLogs.Resource().CopyTo(newResourceLogs.Resource())
		
		// Iterate through scope logs
		for j := 0; j < resourceLogs.ScopeLogs().Len(); j++ {
			scopeLogs := resourceLogs.ScopeLogs().At(j)
			
			// Create a new scope logs entry
			newScopeLogs := newResourceLogs.ScopeLogs().AppendEmpty()
			scopeLogs.Scope().CopyTo(newScopeLogs.Scope())
			
			// Iterate through logs and keep only non-debug logs
			for k := 0; k < scopeLogs.LogRecords().Len(); k++ {
				logRecord := scopeLogs.LogRecords().At(k)
				
				// Check if this is a debug log (severity number <= 5)
				// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#severity-fields
				severityNumber := logRecord.SeverityNumber()
				if severityNumber <= 5 { // Debug or lower
					continue
				}
				
				// Not a debug log, keep it
				newLogRecord := newScopeLogs.LogRecords().AppendEmpty()
				logRecord.CopyTo(newLogRecord)
			}
		}
	}
	
	return filtered
}

// Capabilities returns the capabilities of the processor.
func (p *processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}
