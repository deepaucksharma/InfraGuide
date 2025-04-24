package enhanceddlq

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// The type of the exporter.
	typeStr = "enhanced_dlq"
)

// ErrEmptyConfig is returned when the configuration provided is empty.
var ErrEmptyConfig = errors.New("empty configuration for enhanced_dlq exporter")

// NewFactory creates a new factory for the EnhancedDLQ exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		typeStr,
		CreateDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelAlpha),
		exporter.WithTraces(createTracesExporter, component.StabilityLevelAlpha),
		exporter.WithLogs(createLogsExporter, component.StabilityLevelAlpha),
	)
}

// createMetricsExporter creates a new metrics exporter based on the config.
func createMetricsExporter(
	ctx context.Context,
	set exporter.CreateSettings,
	cfg component.Config,
) (exporter.Metrics, error) {
	eCfg, ok := cfg.(*Config)
	if !ok {
		return nil, ErrEmptyConfig
	}

	exporter, err := newMetricsExporter(ctx, set, eCfg)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetricsExporter(
		ctx,
		set,
		cfg,
		exporter.ConsumeMetrics,
		exporterhelper.WithCapabilities(exporter.Capabilities()),
		exporterhelper.WithTimeout(eCfg.TimeoutSettings),
		exporterhelper.WithQueue(eCfg.QueueSettings),
		exporterhelper.WithRetry(eCfg.RetrySettings),
		exporterhelper.WithStart(exporter.Start),
		exporterhelper.WithShutdown(exporter.Shutdown),
	)
}

// createTracesExporter creates a new traces exporter based on the config.
func createTracesExporter(
	ctx context.Context,
	set exporter.CreateSettings,
	cfg component.Config,
) (exporter.Traces, error) {
	eCfg, ok := cfg.(*Config)
	if !ok {
		return nil, ErrEmptyConfig
	}

	exporter, err := newTracesExporter(ctx, set, eCfg)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		ctx,
		set,
		cfg,
		exporter.ConsumeTraces,
		exporterhelper.WithCapabilities(exporter.Capabilities()),
		exporterhelper.WithTimeout(eCfg.TimeoutSettings),
		exporterhelper.WithQueue(eCfg.QueueSettings),
		exporterhelper.WithRetry(eCfg.RetrySettings),
		exporterhelper.WithStart(exporter.Start),
		exporterhelper.WithShutdown(exporter.Shutdown),
	)
}

// createLogsExporter creates a new logs exporter based on the config.
func createLogsExporter(
	ctx context.Context,
	set exporter.CreateSettings,
	cfg component.Config,
) (exporter.Logs, error) {
	eCfg, ok := cfg.(*Config)
	if !ok {
		return nil, ErrEmptyConfig
	}

	exporter, err := newLogsExporter(ctx, set, eCfg)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewLogsExporter(
		ctx,
		set,
		cfg,
		exporter.ConsumeLogs,
		exporterhelper.WithCapabilities(exporter.Capabilities()),
		exporterhelper.WithTimeout(eCfg.TimeoutSettings),
		exporterhelper.WithQueue(eCfg.QueueSettings),
		exporterhelper.WithRetry(eCfg.RetrySettings),
		exporterhelper.WithStart(exporter.Start),
		exporterhelper.WithShutdown(exporter.Shutdown),
	)
}
