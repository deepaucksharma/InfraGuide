#!/bin/bash
# Status report script for NRDOT+ MVP

set -e

# Function to show usage
usage() {
    echo "NRDOT+ MVP Status Report Script"
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --prometheus      Use Prometheus queries for metrics (default)"
    echo "  --container       Use docker stats for metrics"
    echo "  --json            Output in JSON format"
    echo "  --csv             Output in CSV format"
    echo ""
    exit 1
}

# Parse options
FORMAT="text"
SOURCE="prometheus"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --prometheus) SOURCE="prometheus" ;;
        --container) SOURCE="container" ;;
        --json) FORMAT="json" ;;
        --csv) FORMAT="csv" ;;
        --help|-h) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
    shift
done

# Check dependencies
if [ "$SOURCE" = "prometheus" ]; then
    command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed. Aborting."; exit 1; }
    command -v jq >/dev/null 2>&1 || { echo "jq is required but not installed. Aborting."; exit 1; }
fi

if [ "$SOURCE" = "container" ]; then
    command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting."; exit 1; }
fi

# Function to get metrics from Prometheus
get_prometheus_metrics() {
    # CPU usage
    CPU_USAGE=$(curl -s "http://localhost:9090/api/v1/query?query=rate(process_cpu_seconds_total%7Bservice%3D%22collector%22%7D%5B1m%5D)" | jq -r '.data.result[0].value[1] | tonumber | . * 100' 2>/dev/null || echo "N/A")
    
    # Memory usage
    MEMORY_USAGE=$(curl -s "http://localhost:9090/api/v1/query?query=process_resident_memory_bytes%7Bservice%3D%22collector%22%7D" | jq -r '.data.result[0].value[1] | tonumber / (1024*1024)' 2>/dev/null || echo "N/A")
    
    # APQ fill ratio
    APQ_FILL=$(curl -s "http://localhost:9090/api/v1/query?query=nrdot_mvp_apq_fill_ratio" | jq -r '.data.result[0].value[1] | tonumber' 2>/dev/null || echo "N/A")
    
    # Cardinality limiter metrics
    CL_KEYS=$(curl -s "http://localhost:9090/api/v1/query?query=nrdot_mvp_cardinality_limiter_unique_keysets" | jq -r '.data.result[0].value[1] | tonumber' 2>/dev/null || echo "N/A")
    CL_DROPPED=$(curl -s "http://localhost:9090/api/v1/query?query=sum(nrdot_mvp_cardinality_limiter_dropped_count)" | jq -r '.data.result[0].value[1] | tonumber' 2>/dev/null || echo "N/A")
    
    # DLQ metrics
    DLQ_SIZE=$(curl -s "http://localhost:9090/api/v1/query?query=nrdot_mvp_dlq_size_bytes" | jq -r '.data.result[0].value[1] | tonumber / (1024*1024)' 2>/dev/null || echo "N/A")
    DLQ_FILES=$(curl -s "http://localhost:9090/api/v1/query?query=nrdot_mvp_dlq_files_count" | jq -r '.data.result[0].value[1] | tonumber' 2>/dev/null || echo "N/A")
    
    # ADM level
    ADM_LEVEL=$(curl -s "http://localhost:9090/api/v1/query?query=otelcol_adm_current_level" | jq -r '.data.result[0].value[1] | tonumber' 2>/dev/null || echo "N/A")
}

# Function to get metrics from Docker
get_container_metrics() {
    # CPU usage
    CPU_USAGE=$(docker stats --no-stream --format "{{.CPUPerc}}" nrdot-mvp_collector_1 | sed 's/%//')
    
    # Memory usage
    MEMORY_USAGE=$(docker stats --no-stream --format "{{.MemUsage}}" nrdot-mvp_collector_1 | awk '{print $1}' | sed 's/MiB//')
    
    # Other metrics not available directly from docker stats
    APQ_FILL="N/A (use --prometheus)"
    CL_KEYS="N/A (use --prometheus)"
    CL_DROPPED="N/A (use --prometheus)"
    DLQ_SIZE="N/A (use --prometheus)"
    DLQ_FILES="N/A (use --prometheus)"
    ADM_LEVEL="N/A (use --prometheus)"
}

# Get metrics
if [ "$SOURCE" = "prometheus" ]; then
    get_prometheus_metrics
else
    get_container_metrics
fi

# Output results
if [ "$FORMAT" = "json" ]; then
    cat <<EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "collector": {
    "cpu_usage_percent": $CPU_USAGE,
    "memory_usage_mib": $MEMORY_USAGE,
    "apq_fill_percent": $APQ_FILL,
    "cardinality_limiter": {
      "unique_keysets": $CL_KEYS,
      "dropped_total": $CL_DROPPED
    },
    "dlq": {
      "size_mib": $DLQ_SIZE,
      "files_count": $DLQ_FILES
    },
    "adm_level": $ADM_LEVEL
  }
}
EOF
elif [ "$FORMAT" = "csv" ]; then
    echo "timestamp,cpu_usage_percent,memory_usage_mib,apq_fill_percent,unique_keysets,dropped_total,dlq_size_mib,dlq_files_count,adm_level"
    echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ"),$CPU_USAGE,$MEMORY_USAGE,$APQ_FILL,$CL_KEYS,$CL_DROPPED,$DLQ_SIZE,$DLQ_FILES,$ADM_LEVEL"
else
    echo "NRDOT+ MVP Status Report ($(date))"
    echo "---------------------------------"
    echo "Collector:"
    echo "  CPU Usage:      ${CPU_USAGE}%"
    echo "  Memory Usage:   ${MEMORY_USAGE} MiB"
    echo ""
    echo "APQ:"
    echo "  Fill Ratio:     ${APQ_FILL}%"
    echo ""
    echo "Cardinality Limiter:"
    echo "  Unique Keysets: ${CL_KEYS}"
    echo "  Dropped Total:  ${CL_DROPPED}"
    echo ""
    echo "DLQ:"
    echo "  Size:           ${DLQ_SIZE} MiB"
    echo "  Files:          ${DLQ_FILES}"
    echo ""
    echo "ADM:"
    echo "  Current Level:  ${ADM_LEVEL}"
    echo "---------------------------------"
fi
