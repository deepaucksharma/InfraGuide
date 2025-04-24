#!/bin/bash
# Run a workload test against the NRDOT+ MVP

set -e

echo "Running NRDOT+ MVP workload test..."

# Check if the collector is running
if ! docker ps | grep -q otel-collector; then
  echo "Error: The collector is not running. Start it with 'docker-compose up -d'"
  exit 1
fi

# Run the workload generator with the specified profile
PROFILE=${1:-default}
echo "Using workload profile: $PROFILE"

docker-compose run --rm workload-generator --profile=$PROFILE

echo "Test completed!"
