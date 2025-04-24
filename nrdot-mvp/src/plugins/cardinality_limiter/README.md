# CardinalityLimiter Processor

This processor implements a dynamic cardinality control mechanism for OpenTelemetry data, primarily focusing on metrics.

## Overview

The CardinalityLimiter processor addresses the problem of high-cardinality metrics, which can cause performance issues and increased costs in observability systems. It uses an entropy-based algorithm to identify and limit the number of unique key-sets (combinations of labels/attributes) flowing through the pipeline.

## Features

- Enforces a maximum number of unique key-sets (default: 65,536)
- Uses entropy-based scoring to prioritize important key-sets
- Provides options for dropping or aggregating high-cardinality data
- Can be applied to metrics only, or to all telemetry types
- Exposes metrics for monitoring cardinality and dropped/aggregated data

## Configuration

```yaml
processors:
  cardinality_limiter:
    # Maximum number of unique key-sets allowed
    max_unique_keysets: 65536
    
    # Cardinality control algorithm: "entropy", "lru", or "random"
    algorithm: entropy
    
    # Action on exceeding cardinality: "drop", "aggregate", or "drop_aggregate"
    action: drop_aggregate
    
    # Dimensions to preserve when aggregating
    aggregation_dimensions: ["service.name", "host.name"]
    
    # Whether to apply only to metrics (true) or all telemetry (false)
    metrics_only: true
```

## Implementation Details

The core of the processor is the entropy-based scoring algorithm, which assigns importance scores to different key-sets based on their information content. When the number of unique key-sets exceeds the configured limit, the processor will:

1. Sort all key-sets by their entropy scores
2. Keep the top N most important key-sets (where N is the configured limit)
3. Apply the configured action (drop or aggregate) to the remaining key-sets

The processor uses an efficient hash table implementation to track unique key-sets and their metadata, with O(1) lookups and minimal memory overhead.

## Todo

- [ ] Implement the entropy-based scoring algorithm
- [ ] Add proper metrics for monitoring
- [ ] Add tests for all functionality
- [ ] Document the algorithm in detail
