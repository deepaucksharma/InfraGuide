#!/bin/bash
# Run script for NRDOT+ MVP

set -e

# Function to show usage
usage() {
    echo "NRDOT+ MVP Run Script"
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  up              Start all services"
    echo "  down            Stop all services"
    echo "  logs            View logs"
    echo "  outage [on|off] Simulate outage"
    echo "  storm           Run cardinality storm simulation"
    echo "  status          Show service status"
    echo "  verify          Verify DLQ functionality"
    echo "  help            Show this help"
    echo ""
    echo "Options:"
    echo "  --with-agents   Include agent services (for 'up' command)"
    echo "  --real-nr       Use real New Relic endpoint (requires NEW_RELIC_API_KEY)"
    echo ""
    echo "Environment variables:"
    echo "  NEW_RELIC_API_KEY  New Relic API key for real ingest"
    echo "  NEW_RELIC_ENDPOINT New Relic endpoint URL (default: https://otlp.nr-data.net:4317)"
    echo ""
    exit 1
}

# Check for Docker
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting."; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting."; exit 1; }

# Parse command
if [ $# -eq 0 ]; then
    usage
fi

COMMAND=$1
shift

# Parse options
WITH_AGENTS=false
REAL_NR=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --with-agents) WITH_AGENTS=true ;;
        --real-nr) REAL_NR=true ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
    shift
done

# Set up environment for real New Relic connection
if [ "$REAL_NR" = true ]; then
    if [ -z "$NEW_RELIC_API_KEY" ]; then
        echo "ERROR: --real-nr requires NEW_RELIC_API_KEY environment variable"
        exit 1
    fi
    export NEW_RELIC_ENDPOINT=${NEW_RELIC_ENDPOINT:-https://otlp.nr-data.net:4317}
    echo "Using real New Relic endpoint: $NEW_RELIC_ENDPOINT"
fi

# Create required directories
mkdir -p data/dlq data/prometheus data/grafana

# Execute command
case $COMMAND in
    up)
        echo "Starting NRDOT+ MVP services..."
        if [ "$WITH_AGENTS" = true ]; then
            docker-compose up -d --profile agents
        else
            docker-compose up -d
        fi
        
        # Wait for services to start
        echo "Waiting for services to start..."
        sleep 5
        
        # Print service status
        docker-compose ps
        
        echo "Services started! Access dashboards at:"
        echo "  - Grafana:    http://localhost:3000 (admin/admin)"
        echo "  - Prometheus: http://localhost:9090"
        ;;
        
    down)
        echo "Stopping services..."
        docker-compose down
        ;;
        
    logs)
        docker-compose logs -f
        ;;
        
    outage)
        # Check arguments
        if [ $# -eq 0 ]; then
            echo "Missing argument for outage command. Use: outage [on|off]"
            exit 1
        fi
        
        if [ "$1" = "on" ]; then
            echo "Simulating outage..."
            docker-compose run --rm outage-simulator --target-url=http://mock-service:8080/outage --outage-type=api --duration=300
        elif [ "$1" = "off" ]; then
            echo "Ending outage..."
            curl -X POST http://localhost:8080/outage -H "Content-Type: application/json" -d '{"action":"stop"}'
        else
            echo "Invalid outage argument: $1. Use: outage [on|off]"
            exit 1
        fi
        ;;
        
    storm)
        echo "Running cardinality storm simulation..."
        docker-compose run --rm workload-generator --profile=high_cardinality --duration=60
        ;;
        
    status)
        docker-compose ps
        ;;
        
    verify)
        echo "Verifying DLQ functionality..."
        docker-compose run --rm outage-simulator --verify-dlq
        ;;
        
    help)
        usage
        ;;
        
    *)
        echo "Unknown command: $COMMAND"
        usage
        ;;
esac
