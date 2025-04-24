package cardinalitylimiter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	// The type of the processor.
	typeStr = "cardinality_limiter"
)

// NewFactory creates a new factory for the CardinalityLimiter processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		CreateDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelAlpha),
		processor.WithTraces(createTracesProcessor, component.StabilityLevelAlpha),
		processor.WithLogs(createLogsProcessor, component.StabilityLevelAlpha),
	)
}

// createMetricsProcessor creates a new metrics processor based on the config.
func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	processorConfig := cfg.(*Config)
	return newMetricsProcessor(set.Logger, processorConfig, nextConsumer)
}

// createTracesProcessor creates a new traces processor based on the config.
func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	processorConfig := cfg.(*Config)
	return newTracesProcessor(set.Logger, processorConfig, nextConsumer)
}

// createLogsProcessor creates a new logs processor based on the config.
func createLogsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	processorConfig := cfg.(*Config)
	return newLogsProcessor(set.Logger, processorConfig, nextConsumer)
}
