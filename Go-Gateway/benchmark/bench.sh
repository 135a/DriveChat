#!/bin/bash
# bench.sh - Gateway Performance Benchmark Script
# Usage: ./bench.sh [OPTIONS]
#
# Options:
#   -u URL       Target URL (default: http://localhost:8080)
#   -c CONNS     Number of concurrent connections (default: 100)
#   -d DURATION  Test duration (default: 30s)
#   -t THREADS   Number of threads (default: 4)

set -e

# Defaults
URL="http://localhost:8080"
CONNECTIONS=100
DURATION="30s"
THREADS=4
REPORT_DIR="$(dirname "$0")/reports"

# Parse arguments
while getopts "u:c:d:t:" opt; do
  case $opt in
    u) URL="$OPTARG" ;;
    c) CONNECTIONS="$OPTARG" ;;
    d) DURATION="$OPTARG" ;;
    t) THREADS="$OPTARG" ;;
    *) echo "Usage: $0 [-u url] [-c connections] [-d duration] [-t threads]"; exit 1 ;;
  esac
done

# Create reports directory
mkdir -p "$REPORT_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$REPORT_DIR/bench_${TIMESTAMP}.txt"

echo "============================================"
echo "  Nexus Gateway Performance Benchmark"
echo "============================================"
echo "Target:      $URL"
echo "Connections: $CONNECTIONS"
echo "Duration:    $DURATION"
echo "Threads:     $THREADS"
echo "Report:      $REPORT_FILE"
echo "============================================"
echo ""

# Check wrk is installed
if ! command -v wrk &> /dev/null; then
  echo "Error: wrk is not installed. Please install it first."
  echo "  Ubuntu/Debian: sudo apt-get install wrk"
  echo "  macOS: brew install wrk"
  exit 1
fi

# Run benchmark
echo "Starting benchmark..."
echo ""

wrk -t"$THREADS" -c"$CONNECTIONS" -d"$DURATION" --latency "$URL" | tee "$REPORT_FILE"

echo ""
echo "============================================"
echo "Benchmark complete. Report saved to: $REPORT_FILE"
echo "============================================"
