# NRDOT+ MVP Makefile

# Variables
SHELL := /bin/bash
GO := go
DOCKER := docker
DOCKER_COMPOSE := docker-compose
PROJECT := nrdot-mvp
VERSION := 0.1.0

# Directories
BIN_DIR := bin
DATA_DIR := data
CONFIG_DIR := config
SRC_DIR := src
PLUGINS_DIR := $(SRC_DIR)/plugins

# Build targets
.PHONY: all build clean test lint docker run stop logs up down reload storm outage verify help

all: lint test build

# Build all components
build: plugins collector
	@echo "Build completed successfully"

# Build plugins
plugins:
	@echo "Building plugins..."
	$(GO) build -o $(BIN_DIR)/cardinality_limiter.so $(PLUGINS_DIR)/cardinality_limiter/*.go
	$(GO) build -o $(BIN_DIR)/adaptive_priority_queue.so $(PLUGINS_DIR)/adaptive_priority_queue/*.go
	$(GO) build -o $(BIN_DIR)/enhanced_dlq.so $(PLUGINS_DIR)/enhanced_dlq/*.go

# Build the collector
collector:
	@echo "Building collector..."
	$(GO) build -o $(BIN_DIR)/collector cmd/collector/*.go

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Run linting
lint:
	@echo "Running linters..."
	$(GO) vet ./...
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/*
	rm -rf $(DATA_DIR)/dlq/*
	$(DOCKER_COMPOSE) down --volumes --remove-orphans

# Build Docker images
docker:
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build

# Run the collector locally
run: build
	@echo "Running collector..."
	$(BIN_DIR)/collector --config $(CONFIG_DIR)/collector-config.yaml

# Start all services with docker-compose
up:
	@echo "Starting services..."
	mkdir -p $(DATA_DIR)/dlq
	mkdir -p $(DATA_DIR)/prometheus
	mkdir -p $(DATA_DIR)/grafana
	$(DOCKER_COMPOSE) up -d

# Stop all services
down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

# View logs
logs:
	@echo "Showing logs..."
	$(DOCKER_COMPOSE) logs -f

# Reload services
reload:
	@echo "Reloading services..."
	$(DOCKER_COMPOSE) restart

# Run cardinality storm simulation
storm:
	@echo "Running cardinality storm for 30 seconds..."
	$(DOCKER_COMPOSE) run --rm workload-generator --profile=high_cardinality --duration=30

# Simulate an outage
outage:
	@echo "Simulating outage for 5 minutes..."
	$(DOCKER_COMPOSE) run --rm outage-simulator --duration=300

# Verify DLQ functionality
verify:
	@echo "Verifying DLQ functionality..."
	$(DOCKER_COMPOSE) run --rm outage-simulator --verify-dlq

# Show help
help:
	@echo "NRDOT+ MVP Makefile targets:"
	@echo "  all         - Run lint, test, and build targets"
	@echo "  build       - Build all components"
	@echo "  plugins     - Build plugins only"
	@echo "  collector   - Build collector only"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linters"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker      - Build Docker images"
	@echo "  run         - Run the collector locally"
	@echo "  up          - Start all services with docker-compose"
	@echo "  down        - Stop all services"
	@echo "  logs        - View logs"
	@echo "  reload      - Reload services"
	@echo "  storm       - Run cardinality storm simulation"
	@echo "  outage      - Simulate an outage"
	@echo "  verify      - Verify DLQ functionality"
