# NRDOT+ MVP Running Instructions

Due to the complexity of the dependencies and the need for a specific OpenTelemetry version, the best approach to run this project is using Docker. However, since Docker is not available in this environment, here's a step-by-step guide on how to run the project in a proper environment:

## Prerequisites

- Docker and docker-compose
- Go 1.21 or later (if building locally)
- Git

## Running with Docker (Recommended)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/nrdot-mvp.git
   cd nrdot-mvp
   ```

2. **Run the build script:**
   ```bash
   chmod +x build.sh
   ./build.sh
   ```
   This will build all the necessary components including the plugins and the collector.

3. **Start the services:**
   ```bash
   chmod +x run.sh
   ./run.sh up
   ```
   This will start all the required services using docker-compose.

4. **Run the demo scenarios:**
   ```bash
   # Cardinality storm scenario
   ./scripts/storm.sh
   
   # Outage scenario
   ./scripts/outage.sh on
   # Wait a few minutes then
   ./scripts/outage.sh off
   
   # View system status
   ./scripts/report.sh
   ```

5. **Access the dashboards:**
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090

## Troubleshooting

If you encounter issues with Docker, try these steps:

1. Make sure Docker is running:
   ```bash
   docker info
   ```

2. If you get permission errors, try running with sudo (Linux/Mac):
   ```bash
   sudo ./run.sh up
   ```

3. If you have issues with building the Go components, try:
   ```bash
   go mod tidy
   go build -o bin/collector ./cmd/collector
   ```

## Manual Testing Without Docker

While not recommended due to dependency complexities, you can run some components individually:

1. **Mock Service:**
   ```bash
   go run ./cmd/mock-upstream/main.go
   ```

2. **NR Ingest Mock:**
   ```bash
   go run ./cmd/nr-ingest/main.go
   ```

## Project Structure Overview

- `cmd/` - Main applications (collector, mock-upstream, nr-ingest)
- `src/plugins/` - Custom OpenTelemetry plugins
- `otel-config/` - Configuration files
- `scripts/` - Demo and utility scripts
- `monitoring/` - Prometheus and Grafana configuration
- `docker/` - Docker-related files

## Key Components

1. **CardinalityLimiter** - Implements entropy-based cardinality control
2. **AdaptivePriorityQueue** - Implements WRR scheduling for telemetry
3. **EnhancedDLQ** - Provides durable storage with SHA-256 verification
4. **AdaptiveDegradationManager** - Controls load shedding under pressure

## Next Steps

After running the system, try these activities:

1. Examine how cardinality limiting affects high-cardinality metrics
2. Observe how APQ prioritizes critical telemetry during outages
3. Check DLQ operations during and after outages
4. Monitor system resource usage throughout these tests

For more details, refer to the comprehensive documentation in the `docs/` directory.
