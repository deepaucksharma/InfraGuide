#!/bin/bash
# Cardinality storm simulation script for NRDOT+ MVP

set -e

DURATION=${DURATION:-60}
TAG_COUNT=${TAG_COUNT:-1000000}
RATE=${RATE:-5000}

# Function to show usage
usage() {
    echo "NRDOT+ MVP Cardinality Storm Script"
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --duration=SECONDS  Duration of storm in seconds (default: 60)"
    echo "  --tags=COUNT        Number of unique tags to generate (default: 1000000)"
    echo "  --rate=PER_SECOND   Rate of telemetry items per second (default: 5000)"
    echo ""
    exit 1
}

# Parse options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --duration=*) DURATION="${1#*=}" ;;
        --tags=*) TAG_COUNT="${1#*=}" ;;
        --rate=*) RATE="${1#*=}" ;;
        --help|-h) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
    shift
done

echo "Starting cardinality storm simulation..."
echo "  - Duration: $DURATION seconds"
echo "  - Unique tags: $TAG_COUNT"
echo "  - Rate: $RATE items/second"

# Run workload generator with high cardinality profile
docker-compose run --rm workload-generator --profile=high_cardinality --duration=$DURATION --rate=$RATE

echo "Storm completed! Monitor Grafana dashboards to observe:"
echo "  - CardinalityLimiter dropping/aggregating metrics"
echo "  - CPU usage remaining stable"
echo "  - AdaptiveDegradationManager possibly activating"
