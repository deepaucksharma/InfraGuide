package enhanceddlq

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// metricsExporter is the exporter for metrics.
type metricsExporter struct {
	logger    *zap.Logger
	config    *Config
	storage   *DLQStorage
	forwarder component.Component // This would be the component to forward replayed data to
}

// newMetricsExporter creates a new metrics exporter.
func newMetricsExporter(
	ctx context.Context,
	set exporter.CreateSettings,
	config *Config,
) (*metricsExporter, error) {
	storage, err := NewDLQStorage(config, set.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ storage: %w", err)
	}

	return &metricsExporter{
		logger:  set.Logger,
		config:  config,
		storage: storage,
	}, nil
}

// Start starts the exporter.
func (e *metricsExporter) Start(ctx context.Context, host component.Host) error {
	if e.config.ReplayOnStart {
		return e.StartReplay(ctx)
	}
	return nil
}

// Shutdown stops the exporter.
func (e *metricsExporter) Shutdown(context.Context) error {
	return e.storage.Shutdown()
}

// ConsumeMetrics implements the metrics consumer interface.
func (e *metricsExporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Check if interleaving is active and if we should allow live traffic
	if e.storage.IsReplayActive() && !e.storage.replayInterleave.AllowLive() {
		// Interleaving is active but we should not process live traffic right now
		return nil
	}

	// Serialize metrics to bytes
	serialized, err := serializeMetrics(md)
	if err != nil {
		return fmt.Errorf("failed to serialize metrics: %w", err)
	}

	// Write to DLQ storage
	if err := e.storage.Write(ctx, serialized); err != nil {
		return fmt.Errorf("failed to write metrics to DLQ: %w", err)
	}

	return nil
}

// Capabilities returns the capabilities of the exporter.
func (e *metricsExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// StartReplay starts the replay process.
func (e *metricsExporter) StartReplay(ctx context.Context) error {
	consumer := &metricsReplayConsumer{
		logger:    e.logger,
		forwarder: e.forwarder,
	}
	return e.storage.StartReplay(ctx, consumer)
}

// StopReplay stops the replay process.
func (e *metricsExporter) StopReplay() {
	e.storage.StopReplay()
}

// metricsReplayConsumer implements the DLQConsumer interface for metrics.
type metricsReplayConsumer struct {
	logger    *zap.Logger
	forwarder component.Component
}

// ConsumeDLQRecord implements the DLQConsumer interface.
func (c *metricsReplayConsumer) ConsumeDLQRecord(ctx context.Context, record *DLQRecord) error {
	// Deserialize the metrics
	md, err := deserializeMetrics(record.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize metrics: %w", err)
	}

	// Forward to the next component in the pipeline
	if c.forwarder != nil {
		if consumer, ok := c.forwarder.(consumer.Metrics); ok {
			return consumer.ConsumeMetrics(ctx, md)
		}
	}

	c.logger.Warn("No forwarder configured for metrics replay")
	return nil
}

// serializeMetrics serializes metrics data to bytes.
func serializeMetrics(md pmetric.Metrics) ([]byte, error) {
	// In a real implementation, this would serialize the metrics to a binary format
	// For simplicity, we'll just return a placeholder
	return []byte("serialized_metrics_placeholder"), nil
}

// deserializeMetrics deserializes bytes to metrics data.
func deserializeMetrics(data []byte) (pmetric.Metrics, error) {
	// In a real implementation, this would deserialize the bytes to metrics
	// For simplicity, we'll just return empty metrics
	return pmetric.NewMetrics(), nil
}
