#!/bin/bash
# Outage simulation script for NRDOT+ MVP

set -e

OUTAGE_DURATION=${OUTAGE_DURATION:-300}
OUTAGE_TYPE=${OUTAGE_TYPE:-api}
TARGET_SERVICE=${TARGET_SERVICE:-mock-service}

# Function to show usage
usage() {
    echo "NRDOT+ MVP Outage Simulation Script"
    echo "Usage: $0 [on|off] [--duration=SECONDS] [--type=TYPE]"
    echo ""
    echo "Commands:"
    echo "  on     Start an outage"
    echo "  off    End an outage"
    echo ""
    echo "Options:"
    echo "  --duration=SECONDS   Duration of outage in seconds (default: 300)"
    echo "  --type=TYPE          Type of outage (api, container_stop, network)"
    echo "                       Default: api"
    echo ""
    exit 1
}

# Parse arguments
if [ $# -eq 0 ]; then
    usage
fi

COMMAND=$1
shift

# Parse options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --duration=*) OUTAGE_DURATION="${1#*=}" ;;
        --type=*) OUTAGE_TYPE="${1#*=}" ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
    shift
done

case $COMMAND in
    on)
        echo "Starting $OUTAGE_TYPE outage for $OUTAGE_DURATION seconds..."
        
        case $OUTAGE_TYPE in
            api)
                curl -X POST http://localhost:8080/outage -H "Content-Type: application/json" \
                    -d "{\"action\":\"start\",\"duration_seconds\":$OUTAGE_DURATION}"
                ;;
                
            container_stop)
                echo "Stopping container: $TARGET_SERVICE"
                docker stop $(docker ps -q -f name=$TARGET_SERVICE)
                
                # Auto-restart after duration
                (
                    sleep $OUTAGE_DURATION
                    echo "Auto-restarting container: $TARGET_SERVICE"
                    docker start $(docker ps -aq -f name=$TARGET_SERVICE)
                ) &
                ;;
                
            network)
                echo "Creating network outage using iptables (requires root)..."
                # This requires running the script with sudo
                TARGET_PORT=${TARGET_PORT:-8080}
                
                # Add iptables rule to block traffic
                sudo iptables -A OUTPUT -p tcp --dport $TARGET_PORT -j DROP
                
                # Auto-remove rule after duration
                (
                    sleep $OUTAGE_DURATION
                    echo "Ending network outage..."
                    sudo iptables -D OUTPUT -p tcp --dport $TARGET_PORT -j DROP
                ) &
                ;;
                
            *)
                echo "Unknown outage type: $OUTAGE_TYPE"
                usage
                ;;
        esac
        
        echo "Outage started! Will run for $OUTAGE_DURATION seconds."
        echo "Monitor Grafana dashboards to observe behavior."
        echo "Run '$0 off' to end the outage early."
        ;;
        
    off)
        echo "Ending outage..."
        
        case $OUTAGE_TYPE in
            api)
                curl -X POST http://localhost:8080/outage -H "Content-Type: application/json" \
                    -d '{"action":"stop"}'
                ;;
                
            container_stop)
                echo "Restarting container: $TARGET_SERVICE"
                docker start $(docker ps -aq -f name=$TARGET_SERVICE)
                ;;
                
            network)
                echo "Removing network block (requires root)..."
                TARGET_PORT=${TARGET_PORT:-8080}
                sudo iptables -D OUTPUT -p tcp --dport $TARGET_PORT -j DROP
                ;;
                
            *)
                echo "Unknown outage type: $OUTAGE_TYPE"
                usage
                ;;
        esac
        
        echo "Outage ended! Monitor dashboards to observe recovery."
        ;;
        
    *)
        echo "Unknown command: $COMMAND"
        usage
        ;;
esac
