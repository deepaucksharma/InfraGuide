# EnhancedDLQ Exporter

This exporter implements an enhanced dead-letter queue (DLQ) with file storage, SHA-256 verification, and configurable replay capabilities.

## Overview

The EnhancedDLQ exporter addresses the need for reliable persistence of telemetry data during outages or processing failures. It provides enhanced durability and data integrity guarantees compared to standard DLQ implementations.

## Features

- Persists telemetry data to disk with filesystem fsync
- Uses SHA-256 hashing to verify data integrity
- Supports rate-limited replay of stored data
- Interleaves replay data with live traffic at a configurable ratio
- Supports 72-hour durability with configurable retention
- Exposes metrics for monitoring DLQ size and replay status

## Configuration

```yaml
exporters:
  enhanced_dlq:
    # Directory to store DLQ files
    directory: /var/lib/otel/dlq
    
    # Maximum size of individual DLQ files in MiB
    file_size_limit_mib: 100
    
    # Whether to verify data integrity with SHA-256
    verify_sha256: true
    
    # Maximum replay rate in MiB/s
    replay_rate_mib_sec: 4
    
    # Ratio of replay:live traffic (1 means 1:1)
    interleave_ratio: 1
    
    # Maximum retention period in hours
    retention_hours: 72
```

## Implementation Details

The EnhancedDLQ exporter uses file-based storage with several key features:

1. Data is serialized and written to files with proper fsync for durability
2. SHA-256 hashes are computed and stored alongside the data for integrity verification
3. During replay, data is read at a controlled rate to avoid overwhelming the system
4. Replay is interleaved with live traffic to ensure both are processed
5. A background process manages file rotation, cleanup, and retention policies

The exporter handles various telemetry types (metrics, traces, logs) with appropriate serialization for each.

## Todo

- [ ] Implement the file storage mechanism with proper fsync
- [ ] Add SHA-256 verification for data integrity
- [ ] Implement rate-limited replay functionality
- [ ] Add interleaving of replay and live traffic
- [ ] Add metrics for monitoring DLQ status
- [ ] Add tests for all functionality
