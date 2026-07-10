#!/usr/bin/env bash
set -e

MODE=${1:-lazy}

echo "Mode: $MODE"

echo
echo "Startup latency"

hyperfine \
    --warmup 5 \
    'sudo ./kaiten run /bin/true'

echo
echo "Peak memory"

/usr/bin/time -v \
    sudo ./kaiten run /bin/true