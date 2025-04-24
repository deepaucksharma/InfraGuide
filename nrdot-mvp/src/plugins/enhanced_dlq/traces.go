package enhanceddlq

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// tracesExporter is the exporter for traces.
type tracesExporter struct {
	logger    *zap.Logger
	config    *Config
	storage   *DLQStorage
	forwarder component.Component // This would be the component to forward replayed data to
}

// newTracesExporter creates a new traces exporter.
func newTracesExporter(
	ctx context.Context,
	set exporter.CreateSettings,
	config *Config,
) (*tracesExporter, error) {
	storage, err := NewDLQStorage(config, set.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ storage: %w", err)
	}

	return &tracesExporter{
		logger:  set.Logger,
		config:  config,
		storage: storage,
	}, nil
}

// Start starts the exporter.
func (e *tracesExporter) Start(ctx context.Context, host component.Host) error {
	if e.config.ReplayOnStart {
		return e.StartReplay(ctx)
	}
	return nil
}

// Shutdown stops the exporter.
func (e *tracesExporter) Shutdown(context.Context) error {
	return e.storage.Shutdown()
}

// ConsumeTraces implements the traces consumer interface.
func (e *tracesExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	// Check if interleaving is active and if we should allow live traffic
	if e.storage.IsReplayActive() && !e.storage.replayInterleave.AllowLive() {
		// Interleaving is active but we should not process live traffic right now
		return nil
	}

	// Serialize traces to bytes
	serialized, err := serializeTraces(td)
	if err != nil {
		return fmt.Errorf("failed to serialize traces: %w", err)
	}

	// Write to DLQ storage
	if err := e.storage.Write(ctx, serialized); err != nil {
		return fmt.Errorf("failed to write traces to DLQ: %w", err)
	}

	return nil
}

// Capabilities returns the capabilities of the exporter.
func (e *tracesExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// StartReplay starts the replay process.
func (e *tracesExporter) StartReplay(ctx context.Context) error {
	consumer := &tracesReplayConsumer{
		logger:    e.logger,
		forwarder: e.forwarder,
	}
	return e.storage.StartReplay(ctx, consumer)
}

// StopReplay stops the replay process.
func (e *tracesExporter) StopReplay() {
	e.storage.StopReplay()
}

// tracesReplayConsumer implements the DLQConsumer interface for traces.
type tracesReplayConsumer struct {
	logger    *zap.Logger
	forwarder component.Component
}

// ConsumeDLQRecord implements the DLQConsumer interface.
func (c *tracesReplayConsumer) ConsumeDLQRecord(ctx context.Context, record *DLQRecord) error {
	// Deserialize the traces
	td, err := deserializeTraces(record.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize traces: %w", err)
	}

	// Forward to the next component in the pipeline
	if c.forwarder != nil {
		if consumer, ok := c.forwarder.(consumer.Traces); ok {
			return consumer.ConsumeTraces(ctx, td)
		}
	}

	c.logger.Warn("No forwarder configured for traces replay")
	return nil
}

// serializeTraces serializes traces data to bytes.
func serializeTraces(td ptrace.Traces) ([]byte, error) {
	// In a real implementation, this would serialize the traces to a binary format
	// For simplicity, we'll just return a placeholder
	return []byte("serialized_traces_placeholder"), nil
}

// deserializeTraces deserializes bytes to traces data.
func deserializeTraces(data []byte) (ptrace.Traces, error) {
	// In a real implementation, this would deserialize the bytes to traces
	// For simplicity, we'll just return empty traces
	return ptrace.NewTraces(), nil
}
