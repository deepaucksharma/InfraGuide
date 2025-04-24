# NRDOT+ MVP

This repository contains a proof-of-concept implementation of the NRDOT+ v9.0 platform, focusing on three key technical capabilities:

1. **Dynamic cardinality control** - Entropy-based cardinality limiting to 500k unique key-sets
2. **Priority queuing with spilling to disk** - Adaptive Priority Queue with 5:3:1 WRR scheduling
3. **Enhanced durability and resilience** - 72-hour durability with SHA-256 verification and replay

## Overview

The NRDOT+ MVP demonstrates these capabilities with extremely low resource usage (≤2% CPU, ≤150 MiB memory) while handling significant telemetry volumes. It's designed to be run locally with Docker or in a Kubernetes environment.

## Architecture

The MVP consists of these key components:

1. **OpenTelemetry Collector** with three experimental plugins:
   - CardinalityLimiter processor (entropy-based)
   - Adaptive Priority Queue (APQ) wrapper (3 classes, WRR)
   - Enhanced DLQ (file-storage with SHA-256)

2. **External Agents**:
   - nr-ebpf-integration for Linux kernel exec/net flow capture
   - nr-profiler-integration for continuous profiling

3. **Testing Components**:
   - mock-upstream service for simulating outages and errors
   - nr-ingest mock service for simulating New Relic backend

4. **Monitoring**:
   - Prometheus for metrics collection
   - Grafana for visualization

## Getting Started

### Prerequisites

- Docker and docker-compose
- Go 1.21+ (for building from source)
- Linux environment (for eBPF agent functionality)

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/your-org/nrdot-mvp.git
   cd nrdot-mvp
   ```

2. Build the components:
   ```bash
   ./build.sh
   ```

3. Run the services:
   ```bash
   ./run.sh up
   ```

4. Access the dashboards:
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090

### Optional: Connect to Real New Relic

To use a real New Relic account instead of the mock ingest service:

```bash
export NEW_RELIC_API_KEY=your_license_key_here
./run.sh up --real-nr
```

## Demonstration Scenarios

The repository includes scripts to demonstrate key capabilities:

### 1. Cardinality Control

Run a cardinality storm that generates millions of unique tag combinations:

```bash
./scripts/storm.sh
```

Observe in Grafana:
- Cardinals Limiter dropping/aggregating high-cardinality metrics
- CPU usage remaining stable despite high cardinality load

### 2. Resilience with DLQ

Simulate a backend outage and observe data preservation:

```bash
# Start an outage
./scripts/outage.sh on

# Wait a few minutes, then end the outage
./scripts/outage.sh off
```

Observe in Grafana:
- Data spilling to DLQ during outage
- APQ fill percentage increasing
- Automatic replay after outage ends

### 3. System Status

Check the current status of the system:

```bash
./scripts/report.sh
```

## Component Details

### CardinalityLimiter

The CardinalityLimiter uses an entropy-based algorithm to determine which key-sets to keep, drop, or aggregate:

- Fixed 500k open-address hash map (FNV-1a 64)
- Entropy-based scoring of label sets
- Configurable thresholds (0.75, 0.9) for aggregation/dropping

### Adaptive Priority Queue (APQ)

The APQ implements weighted round-robin scheduling with three priority levels:

- O(1) enqueue/dequeue operations
- WRR scheduling with 5:3:1 ratio (critical:high:normal)
- Circuit breaker pattern for backend issues
- DLQ spill when queue exceeds 95% capacity

### Enhanced DLQ

The Enhanced DLQ provides durable storage with data integrity verification:

- 128 MiB segment files with metadata headers
- zstd-3 compression for efficient storage
- SHA-256 verification for data integrity
- Token bucket rate limiter for replay (4 MiB/s)
- 1:1 interleaving of replay and live traffic

### WASM Processor

The MVP includes a WebAssembly processor for data transformation:

- PII masking (credit cards, SSNs, passwords)
- Low overhead (< 5 µs per record)
- Hot-reloadable modules

## Configuration

The collector configuration is in `otel-config/collector.yaml`. Key settings include:

- CardinalityLimiter thresholds and aggregation rules
- APQ priority classes and weights
- DLQ storage location and replay settings
- WASM module configuration

## Performance Targets

The system is designed to meet these performance targets:

- **CPU Usage**: ≤2% on m6i.large (or equivalent)
- **Memory Usage**: ≤150 MiB RSS (including 64 MiB ballast)
- **Latency**: P99 pipeline ≤50 ms
- **Durability**: 24h at 100k items/h → 15 GiB DLQ

## License

This project is licensed under the MIT License - see the LICENSE file for details.
