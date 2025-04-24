#!/bin/bash
# Development environment setup script for NRDOT+ MVP

set -e

echo "Setting up NRDOT+ MVP development environment..."

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting."; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting."; exit 1; }

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Create necessary directories
mkdir -p .build
mkdir -p .data/dlq
mkdir -p .data/prometheus
mkdir -p .data/grafana

# Build Docker images
echo "Building Docker images..."
docker-compose build

echo "Setup complete! You can now run 'docker-compose up' to start the environment."
