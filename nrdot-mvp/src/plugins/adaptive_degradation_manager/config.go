package adaptivedegradationmanager

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
)

// DegradationLevel represents a degradation level with specific actions
type DegradationLevel struct {
	ID      int      `mapstructure:"id"`
	Actions []string `mapstructure:"actions"`
}

// Triggers defines the conditions that trigger degradation levels
type Triggers struct {
	MemoryUtilizationHigh int `mapstructure:"memory_utilization_high"`
	QueueUtilizationHigh  int `mapstructure:"queue_utilization_high"`
	CPUUtilizationHigh    int `mapstructure:"cpu_utilization_high"`
	LatencyP99High        int `mapstructure:"latency_p99_high"`
	ErrorRateHigh         int `mapstructure:"error_rate_high"`
}

// Config defines the configuration for the AdaptiveDegradationManager processor.
type Config struct {
	// Triggers that lead to degradation level changes
	Triggers Triggers `mapstructure:"triggers"`

	// Degradation levels with associated actions
	Levels []DegradationLevel `mapstructure:"levels"`

	// How often to check conditions (in seconds)
	CheckInterval int `mapstructure:"check_interval"`

	// How long to wait before reducing degradation level (in seconds)
	CooldownPeriod int `mapstructure:"cooldown_period"`
}

// Validate validates the processor configuration.
func (cfg *Config) Validate() error {
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = 5
	}

	if cfg.CooldownPeriod <= 0 {
		cfg.CooldownPeriod = 60
	}

	// Ensure we have at least one degradation level
	if len(cfg.Levels) == 0 {
		return fmt.Errorf("at least one degradation level must be configured")
	}

	// Validate actions in each level
	validActions := map[string]bool{
		"inc_batch":       true,
		"stretch_scrape":  true,
		"enable_sampling": true,
		"drop_debug":      true,
		"drop_metrics":    true,
	}

	for _, level := range cfg.Levels {
		for _, action := range level.Actions {
			if !validActions[action] {
				return fmt.Errorf("invalid action '%s' in degradation level %d", action, level.ID)
			}
		}
	}

	// Ensure triggers are reasonable
	if cfg.Triggers.MemoryUtilizationHigh <= 0 {
		cfg.Triggers.MemoryUtilizationHigh = 75
	} else if cfg.Triggers.MemoryUtilizationHigh > 95 {
		return fmt.Errorf("memory_utilization_high must be <= 95")
	}

	if cfg.Triggers.QueueUtilizationHigh <= 0 {
		cfg.Triggers.QueueUtilizationHigh = 70
	} else if cfg.Triggers.QueueUtilizationHigh > 95 {
		return fmt.Errorf("queue_utilization_high must be <= 95")
	}

	if cfg.Triggers.CPUUtilizationHigh <= 0 {
		cfg.Triggers.CPUUtilizationHigh = 80
	}

	if cfg.Triggers.LatencyP99High <= 0 {
		cfg.Triggers.LatencyP99High = 500
	}

	if cfg.Triggers.ErrorRateHigh <= 0 {
		cfg.Triggers.ErrorRateHigh = 10
	} else if cfg.Triggers.ErrorRateHigh > 100 {
		return fmt.Errorf("error_rate_high must be <= 100")
	}

	return nil
}

// CreateDefaultConfig creates the default configuration for the processor.
func CreateDefaultConfig() component.Config {
	return &Config{
		Triggers: Triggers{
			MemoryUtilizationHigh: 75,
			QueueUtilizationHigh:  70,
			CPUUtilizationHigh:    80,
			LatencyP99High:        500,
			ErrorRateHigh:         10,
		},
		Levels: []DegradationLevel{
			{
				ID:      1,
				Actions: []string{"inc_batch", "stretch_scrape"},
			},
			{
				ID:      2,
				Actions: []string{"enable_sampling"},
			},
			{
				ID:      3,
				Actions: []string{"drop_debug", "drop_metrics"},
			},
		},
		CheckInterval:  5,
		CooldownPeriod: 60,
	}
}
