package cardinalitylimiter

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration for the CardinalityLimiter processor.
type Config struct {
	// MaxUniqueKeySets is the maximum number of unique key sets allowed in the hash table.
	// Default: 65536
	MaxUniqueKeySets int `mapstructure:"max_unique_keysets"`

	// Algorithm defines the cardinality control algorithm to use.
	// Options: "entropy", "lru", "random"
	// Default: "entropy"
	Algorithm string `mapstructure:"algorithm"`

	// Action defines what happens when cardinality exceeds the limit.
	// Options: "drop", "aggregate", "drop_aggregate"
	// Default: "drop_aggregate"
	Action string `mapstructure:"action"`

	// AggregationDimensions defines the dimensions to preserve when aggregating.
	// Only used when Action is "aggregate" or "drop_aggregate".
	AggregationDimensions []string `mapstructure:"aggregation_dimensions"`

	// MetricsOnly indicates whether to apply cardinality control only to metrics.
	// If false, the processor will also analyze and limit trace and log attributes.
	// Default: true
	MetricsOnly bool `mapstructure:"metrics_only"`
}

// Validate validates the processor configuration.
func (cfg *Config) Validate() error {
	if cfg.MaxUniqueKeySets <= 0 {
		cfg.MaxUniqueKeySets = 65536
	}

	if cfg.Algorithm == "" {
		cfg.Algorithm = "entropy"
	}

	if cfg.Action == "" {
		cfg.Action = "drop_aggregate"
	}

	return nil
}

// CreateDefaultConfig creates the default configuration for the processor.
func CreateDefaultConfig() component.Config {
	return &Config{
		MaxUniqueKeySets:      65536,
		Algorithm:             "entropy",
		Action:                "drop_aggregate",
		AggregationDimensions: []string{"service.name", "host.name"},
		MetricsOnly:           true,
	}
}
