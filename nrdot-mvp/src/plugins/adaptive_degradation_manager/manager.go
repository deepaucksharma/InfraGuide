package adaptivedegradationmanager

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// newProcessor creates a generic processor for metrics, traces and logs
func newProcessor(logger *zap.Logger, config *Config, nextConsumer interface{}) (*processor, error) {
	p := &processor{
		logger:          logger,
		config:          config,
		currentLevel:    0,
		lastLevelChange: 0,
		sampleRate:      1.0,
		batchMultiplier: 1,
		scrapeMultiplier: 1,
		dropDebug:       false,
		dropMetrics:     false,
	}
	
	// Set up metrics
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

	// Store the appropriate consumer based on type
	switch c := nextConsumer.(type) {
	case consumer.Traces:
		p.nextTraceConsumer = c
	case consumer.Metrics:
		p.nextMetricConsumer = c
	case consumer.Logs:
		p.nextLogConsumer = c
	default:
		logger.Error("Unsupported consumer type")
	}

	return p, nil
}

// processor implements metrics/traces/logs consumer interfaces
type processor struct {
	logger             *zap.Logger
	config             *Config
	nextMetricConsumer consumer.Metrics
	nextTraceConsumer  consumer.Traces
	nextLogConsumer    consumer.Logs
	
	// State
	currentLevel      int32
	lastLevelChange   int64
	stateMutex        sync.RWMutex
	
	// Metrics tracking
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
}

// Start starts the processor
func (p *processor) Start(ctx context.Context, host component.Host) error {
	// Start monitoring stats in the background
	return nil
}

// Shutdown stops the processor
func (p *processor) Shutdown(ctx context.Context) error {
	// Clean up resources
	return nil
}

// ConsumeTraces implements the consumer.Traces interface
func (p *processor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if p.nextTraceConsumer == nil {
		return nil
	}
	return p.nextTraceConsumer.ConsumeTraces(ctx, td)
}

// ConsumeMetrics implements the consumer.Metrics interface  
func (p *processor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if p.nextMetricConsumer == nil {
		return nil
	}
	return p.nextMetricConsumer.ConsumeMetrics(ctx, md)
}

// ConsumeLogs implements the consumer.Logs interface
func (p *processor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	if p.nextLogConsumer == nil {
		return nil
	}
	return p.nextLogConsumer.ConsumeLogs(ctx, ld)
}

// Capabilities returns the consumer capabilities
func (p *processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}
