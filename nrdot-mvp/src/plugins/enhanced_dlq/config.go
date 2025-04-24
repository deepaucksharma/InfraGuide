package enhanceddlq

import (
	"path/filepath"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines the configuration for the EnhancedDLQ exporter.
type Config struct {
	// Directory is the path to store DLQ files
	Directory string `mapstructure:"directory"`

	// FileSizeLimitMiB is the maximum size of individual DLQ files in MiB
	FileSizeLimitMiB int `mapstructure:"file_size_limit_mib"`

	// VerifySHA256 enables SHA-256 verification for data integrity
	VerifySHA256 bool `mapstructure:"verify_sha256"`

	// ReplayRateMiBSec is the maximum replay rate in MiB/s
	ReplayRateMiBSec float64 `mapstructure:"replay_rate_mib_sec"`

	// InterleaveRatio controls the ratio of replay:live traffic (1 means 1:1)
	InterleaveRatio int `mapstructure:"interleave_ratio"`

	// RetentionHours is the maximum retention period in hours
	RetentionHours int `mapstructure:"retention_hours"`

	// FilePrefix is the prefix for DLQ files
	FilePrefix string `mapstructure:"file_prefix"`

	// ReplayOnStart indicates whether to automatically replay DLQ on startup
	ReplayOnStart bool `mapstructure:"replay_on_start"`

	// ReplayConcurrency is the number of goroutines used for replay
	ReplayConcurrency int `mapstructure:"replay_concurrency"`

	// Common exporter settings
	exporterhelper.TimeoutSettings `mapstructure:",squash"`
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`
}

// Validate validates the exporter configuration.
func (cfg *Config) Validate() error {
	// Validate Directory
	if cfg.Directory == "" {
		cfg.Directory = "/var/lib/otel/dlq"
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(cfg.Directory)
	if err == nil {
		cfg.Directory = absPath
	}

	// Validate FileSizeLimitMiB
	if cfg.FileSizeLimitMiB <= 0 {
		cfg.FileSizeLimitMiB = 100
	}

	// Validate ReplayRateMiBSec
	if cfg.ReplayRateMiBSec <= 0 {
		cfg.ReplayRateMiBSec = 4
	}

	// Validate InterleaveRatio
	if cfg.InterleaveRatio <= 0 {
		cfg.InterleaveRatio = 1
	}

	// Validate RetentionHours
	if cfg.RetentionHours <= 0 {
		cfg.RetentionHours = 72
	}

	// Validate FilePrefix
	if cfg.FilePrefix == "" {
		cfg.FilePrefix = "otel-dlq"
	}

	// Validate ReplayConcurrency
	if cfg.ReplayConcurrency <= 0 {
		cfg.ReplayConcurrency = 1
	}

	return nil
}

// CreateDefaultConfig creates the default configuration for the exporter.
func CreateDefaultConfig() component.Config {
	return &Config{
		Directory:         "/var/lib/otel/dlq",
		FileSizeLimitMiB:  100,
		VerifySHA256:      true,
		ReplayRateMiBSec:  4,
		InterleaveRatio:   1,
		RetentionHours:    72,
		FilePrefix:        "otel-dlq",
		ReplayOnStart:     false,
		ReplayConcurrency: 1,
		TimeoutSettings:   exporterhelper.NewDefaultTimeoutSettings(),
		QueueSettings:     exporterhelper.NewDefaultQueueSettings(),
		RetrySettings:     exporterhelper.NewDefaultRetrySettings(),
	}
}
