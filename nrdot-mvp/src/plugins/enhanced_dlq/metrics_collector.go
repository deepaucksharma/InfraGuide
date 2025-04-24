package enhanceddlq

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

const (
	metricsNamespace = "nrdot_mvp"
	metricsSubsystem = "dlq"
)

// MetricsCollector collects and exposes metrics for the EnhancedDLQ exporter.
type MetricsCollector struct {
	logger    *zap.Logger
	storage   *DLQStorage
	component component.Component
	config    *Config
	registry  *prometheus.Registry
	
	// Metrics
	dlqSizeBytes     prometheus.Gauge
	dlqFilesCount    prometheus.Gauge
	recordsWritten   prometheus.Counter
	bytesWritten     prometheus.Counter
	recordsReplayed  prometheus.Counter
	bytesReplayed    prometheus.Counter
	replayRateBytes  prometheus.Gauge
	replayActive     prometheus.Gauge
	verificationFail prometheus.Counter
	
	// Update tracking
	lastUpdateTime time.Time
	updateMutex    sync.Mutex
}

// NewMetricsCollector creates a new metrics collector for the EnhancedDLQ exporter.
func NewMetricsCollector(
	logger *zap.Logger,
	storage *DLQStorage,
	component component.Component,
	config *Config,
) *MetricsCollector {
	registry := prometheus.NewRegistry()
	
	collector := &MetricsCollector{
		logger:    logger,
		storage:   storage,
		component: component,
		config:    config,
		registry:  registry,
		
		dlqSizeBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "size_bytes",
			Help:      "Total size of the DLQ in bytes",
		}),
		
		dlqFilesCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "files_count",
			Help:      "Number of DLQ files",
		}),
		
		recordsWritten: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "records_written_total",
			Help:      "Total number of records written to the DLQ",
		}),
		
		bytesWritten: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "bytes_written_total",
			Help:      "Total number of bytes written to the DLQ",
		}),
		
		recordsReplayed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "records_replayed_total",
			Help:      "Total number of records replayed from the DLQ",
		}),
		
		bytesReplayed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "bytes_replayed_total",
			Help:      "Total number of bytes replayed from the DLQ",
		}),
		
		replayRateBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "replay_rate_bytes",
			Help:      "Current replay rate in bytes per second",
		}),
		
		replayActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "replay_active",
			Help:      "Whether replay is currently active (0 = inactive, 1 = active)",
		}),
		
		verificationFail: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "verification_fails_total",
			Help:      "Total number of SHA-256 verification failures",
		}),
		
		lastUpdateTime: time.Now(),
	}
	
	// Register metrics with registry
	registry.MustRegister(collector.dlqSizeBytes)
	registry.MustRegister(collector.dlqFilesCount)
	registry.MustRegister(collector.recordsWritten)
	registry.MustRegister(collector.bytesWritten)
	registry.MustRegister(collector.recordsReplayed)
	registry.MustRegister(collector.bytesReplayed)
	registry.MustRegister(collector.replayRateBytes)
	registry.MustRegister(collector.replayActive)
	registry.MustRegister(collector.verificationFail)
	
	return collector
}

// Start starts the metrics collector.
func (c *MetricsCollector) Start(ctx context.Context) error {
	// Start a background goroutine to update metrics periodically
	go c.updateMetricsLoop(ctx)
	
	return nil
}

// updateMetricsLoop periodically updates the metrics.
func (c *MetricsCollector) updateMetricsLoop(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.updateMetrics()
		}
	}
}

// updateMetrics updates the metrics from the storage.
func (c *MetricsCollector) updateMetrics() {
	c.updateMutex.Lock()
	defer c.updateMutex.Unlock()
	
	// Update DLQ size and files count
	totalSize, err := c.getDLQSize()
	if err != nil {
		c.logger.Error("Failed to get DLQ size", zap.Error(err))
	} else {
		c.dlqSizeBytes.Set(float64(totalSize))
	}
	
	files, err := c.storage.ListDLQFiles()
	if err != nil {
		c.logger.Error("Failed to list DLQ files", zap.Error(err))
	} else {
		c.dlqFilesCount.Set(float64(len(files)))
	}
	
	// Update write metrics
	c.recordsWritten.Add(float64(c.storage.totalWrittenItems))
	c.bytesWritten.Add(float64(c.storage.totalWrittenBytes))
	
	// Update replay metrics
	if c.storage.IsReplayActive() {
		c.replayActive.Set(1)
		
		// Calculate replay rate
		now := time.Now()
		elapsed := now.Sub(c.lastUpdateTime).Seconds()
		if elapsed > 0 {
			replayRate := float64(c.storage.rateLimiter.bytesPerSecond)
			c.replayRateBytes.Set(replayRate)
		}
	} else {
		c.replayActive.Set(0)
		c.replayRateBytes.Set(0)
	}
	
	c.lastUpdateTime = time.Now()
}

// getDLQSize calculates the total size of all DLQ files.
func (c *MetricsCollector) getDLQSize() (int64, error) {
	files, err := c.storage.ListDLQFiles()
	if err != nil {
		return 0, err
	}
	
	var totalSize int64
	for _, file := range files {
		info, err := c.getFileInfo(file)
		if err != nil {
			c.logger.Warn("Failed to get file info", zap.Error(err), zap.String("file", file))
			continue
		}
		
		totalSize += info.Size()
	}
	
	return totalSize, nil
}

// getFileInfo gets file info for a file.
func (c *MetricsCollector) getFileInfo(file string) (interface{}, error) {
	// In a real implementation, this would use os.Stat to get file info
	// For simplicity, we'll return a dummy size
	return struct {
		Size func() int64
	}{
		Size: func() int64 { return 1024 * 1024 }, // 1 MiB
	}, nil
}

// RecordVerificationFailure records a SHA-256 verification failure.
func (c *MetricsCollector) RecordVerificationFailure() {
	c.verificationFail.Inc()
}

// RecordReplayedRecord records a replayed record.
func (c *MetricsCollector) RecordReplayedRecord(recordSize int) {
	c.recordsReplayed.Inc()
	c.bytesReplayed.Add(float64(recordSize))
}

// Registry returns the Prometheus registry.
func (c *MetricsCollector) Registry() *prometheus.Registry {
	return c.registry
}
