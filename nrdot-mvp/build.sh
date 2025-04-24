#!/bin/bash
# Build script for NRDOT+ MVP

set -e

echo "Building NRDOT+ MVP..."

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting."; exit 1; }

# Build directories
mkdir -p bin plugins/bin .build data/dlq data/prometheus data/grafana

# Build Go plugins
echo "Building plugins..."
for plugin in cardinality_limiter adaptive_priority_queue enhanced_dlq adaptive_degradation_manager; do
    echo "  - Building plugin: $plugin"
    go build -buildmode=plugin -o plugins/$plugin.so ./src/plugins/$plugin/*.go
done

# Build collector
echo "Building collector..."
go build -o bin/collector ./cmd/collector

# Build mock-upstream
echo "Building mock-upstream..."
go build -o bin/mock-upstream ./cmd/mock-upstream

# Build nr-ingest
echo "Building nr-ingest..."
go build -o bin/nr-ingest ./cmd/nr-ingest

# Verify plugins and permissions
echo "Verifying plugins and permissions..."
chmod 755 plugins/*.so
chmod 755 plugins/pii_masker.wasm

echo "Build completed successfully!"
echo "You can now run: ./run.sh up"
