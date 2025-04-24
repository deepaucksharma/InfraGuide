package adaptivepriorityqueue

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration for the AdaptivePriorityQueue processor.
type Config struct {
	// Priorities defines the weights for each priority level.
	// The keys are the priority level names, and the values are the weights.
	// Default: critical=5, high=3, normal=1
	Priorities map[string]int `mapstructure:"priorities"`

	// MaxQueueSize is the maximum number of items that can be held in the queue.
	// Default: 10000
	MaxQueueSize int `mapstructure:"max_queue_size"`

	// QueueFullThreshold is the percentage of the queue that, when reached,
	// triggers the overflow strategy. Value should be between 0 and 100.
	// Default: 95
	QueueFullThreshold int `mapstructure:"queue_full_threshold"`

	// OverflowStrategy defines what happens when the queue is full.
	// Options: "drop", "dlq", "block"
	// Default: "dlq"
	OverflowStrategy string `mapstructure:"overflow_strategy"`

	// CircuitBreakerEnabled enables the circuit breaker to detect backend issues.
	// Default: true
	CircuitBreakerEnabled bool `mapstructure:"circuit_breaker_enabled"`

	// CircuitBreakerErrorThreshold is the error percentage threshold for tripping the circuit.
	// Default: 50
	CircuitBreakerErrorThreshold int `mapstructure:"circuit_breaker_error_threshold"`

	// CircuitBreakerResetTimeout is the time in seconds after which to try closing the circuit.
	// Default: 60
	CircuitBreakerResetTimeout int `mapstructure:"circuit_breaker_reset_timeout"`
}

// Validate validates the processor configuration.
func (cfg *Config) Validate() error {
	// Set default priorities if not specified
	if len(cfg.Priorities) == 0 {
		cfg.Priorities = map[string]int{
			"critical": 5,
			"high":     3,
			"normal":   1,
		}
	}

	// Set default max queue size if not specified
	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = 10000
	}

	// Set default queue full threshold if not specified or invalid
	if cfg.QueueFullThreshold <= 0 || cfg.QueueFullThreshold > 100 {
		cfg.QueueFullThreshold = 95
	}

	// Set default overflow strategy if not specified
	if cfg.OverflowStrategy == "" {
		cfg.OverflowStrategy = "dlq"
	}

	// Set default circuit breaker error threshold if not specified or invalid
	if cfg.CircuitBreakerErrorThreshold <= 0 || cfg.CircuitBreakerErrorThreshold > 100 {
		cfg.CircuitBreakerErrorThreshold = 50
	}

	// Set default circuit breaker reset timeout if not specified
	if cfg.CircuitBreakerResetTimeout <= 0 {
		cfg.CircuitBreakerResetTimeout = 60
	}

	return nil
}

// CreateDefaultConfig creates the default configuration for the processor.
func CreateDefaultConfig() component.Config {
	return &Config{
		Priorities: map[string]int{
			"critical": 5,
			"high":     3,
			"normal":   1,
		},
		MaxQueueSize:                10000,
		QueueFullThreshold:          95,
		OverflowStrategy:            "dlq",
		CircuitBreakerEnabled:       true,
		CircuitBreakerErrorThreshold: 50,
		CircuitBreakerResetTimeout:   60,
	}
}
