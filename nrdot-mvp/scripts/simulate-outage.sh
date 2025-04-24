#!/bin/bash
# Simulate an outage in the NRDOT+ MVP environment

set -e

echo "Simulating outage in NRDOT+ MVP environment..."

# Check if the services are running
if ! docker ps | grep -q otel-collector; then
  echo "Error: The collector is not running. Start it with 'docker-compose up -d'"
  exit 1
fi

if ! docker ps | grep -q mock-service; then
  echo "Error: The mock service is not running. Start it with 'docker-compose up -d'"
  exit 1
fi

# Get the duration of the outage (default: 30 seconds)
DURATION=${1:-30}
echo "Simulating a $DURATION second outage..."

# Stop the mock service to simulate backend outage
echo "Stopping mock service..."
docker-compose stop mock-service

# Wait for the specified duration
echo "Waiting for $DURATION seconds..."
sleep $DURATION

# Restart the mock service
echo "Restarting mock service..."
docker-compose start mock-service

echo "Outage simulation completed! Check the logs and metrics to verify resilience."
