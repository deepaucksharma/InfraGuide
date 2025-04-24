package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/exporter/prometheusexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"

	// Import custom components
	"github.com/yourusername/nrdot-mvp/src/plugins/adaptive_priority_queue"
	"github.com/yourusername/nrdot-mvp/src/plugins/cardinality_limiter"
	"github.com/yourusername/nrdot-mvp/src/plugins/enhanced_dlq"
	"github.com/yourusername/nrdot-mvp/src/plugins/adaptive_degradation_manager"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	factories, err := components()
	if err != nil {
		logger.Fatal("Failed to build components", zap.Error(err))
		os.Exit(1)
	}

	// Get the config file path from environment or use default
	configPath := os.Getenv("OTEL_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/otel/config.yaml"
	}

	info := component.BuildInfo{
		Command:     "nrdot-collector",
		Description: "NRDOT+ MVP OpenTelemetry Collector",
		Version:     "0.1.0",
	}

	params := service.CollectorSettings{
		BuildInfo: info,
		Factories: factories,
		LoggingOptions: []zap.Option{
			zap.Fields(zap.String("service", "nrdot-collector")),
		},
		ConfigProviderSettings: service.ConfigProviderSettings{
			ConfigMapProvider: confmap.ProviderSettings{
				URIs:      []string{fmt.Sprintf("file:%s", configPath)},
			},
		},
	}

	if err := service.RunAndWaitForShutdown(ctx, params, logger); err != nil {
		logger.Fatal("Application run finished with error", zap.Error(err))
		os.Exit(1)
	}
}

func components() (otelcol.Factories, error) {
	factories := otelcol.Factories{
		Extensions: map[component.Type]extension.Factory{},
		Receivers: map[component.Type]receiver.Factory{
			"otlp": otlpreceiver.NewFactory(),
		},
		Processors: map[component.Type]processor.Factory{
			"batch":                    batchprocessor.NewFactory(),
			"memory_limiter":           memorylimiterprocessor.NewFactory(),
			"cardinality_limiter":      cardinalitylimiter.NewFactory(),
			"adaptive_priority_queue":  adaptivepriorityqueue.NewFactory(),
			"adaptiveDegradationManager": adaptivedegradationmanager.NewFactory(),
		},
		Exporters: map[component.Type]exporter.Factory{
			"otlp":         otlpexporter.NewFactory(),
			"otlphttp":     otlphttpexporter.NewFactory(),
			"prometheus":   prometheusexporter.NewFactory(),
			"enhanced_dlq": enhanceddlq.NewFactory(),
		},
	}

	return factories, nil
}
