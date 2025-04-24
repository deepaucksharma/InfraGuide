# AdaptivePriorityQueue Processor

This processor implements an adaptive priority queue with weighted round-robin scheduling for OpenTelemetry data.

## Overview

The AdaptivePriorityQueue processor addresses the need for priority-based processing of telemetry data. It ensures that critical data gets processed preferentially while still allowing lower-priority data to flow through the system at a controlled rate.

## Features

- Implements weighted round-robin scheduling with configurable weights (default 5:3:1)
- Supports multiple priority levels (critical, high, normal)
- Spills to DLQ when the queue is near capacity (≥95%)
- Includes circuit breaker pattern to detect backend issues
- Exposes metrics for monitoring queue status and throughput

## Configuration

```yaml
processors:
  adaptive_priority_queue:
    # Priority weights for WRR scheduling
    priorities:
      critical: 5
      high: 3
      normal: 1
    
    # Maximum queue size
    max_queue_size: 10000
    
    # Threshold (percentage) at which to trigger overflow strategy
    queue_full_threshold: 95
    
    # Strategy when queue is full: "drop", "dlq", or "block"
    overflow_strategy: dlq
    
    # Circuit breaker settings
    circuit_breaker_enabled: true
    circuit_breaker_error_threshold: 50
    circuit_breaker_reset_timeout: 60
```

## Implementation Details

The adaptive priority queue uses a combination of a priority queue data structure and a weighted round-robin scheduling algorithm:

1. Incoming telemetry is assigned a priority based on its content and metadata
2. Items are enqueued in a priority queue data structure
3. The dequeue operation uses WRR to select which priority level to dequeue next
4. When the queue exceeds the configured threshold, the overflow strategy is applied
5. The circuit breaker monitors success/failure rates and trips when errors exceed the threshold

The WRR scheduling ensures that even during high load, critical data gets processed at a higher rate while still allowing some lower-priority data through.

## Todo

- [ ] Complete the priority determination algorithm
- [ ] Implement the DLQ overflow handler
- [ ] Add proper metrics for monitoring
- [ ] Add tests for all functionality
- [ ] Document the WRR algorithm in detail
