#!/usr/bin/env bash
set -e

ROOT=/tmp/kaiten-bench

rm -rf "$ROOT"
mkdir -p "$ROOT/data"

echo "Generating 5 GB dataset..."

for i in $(seq 1 5000); do
    dd if=/dev/urandom \
       of="$ROOT/data/$i.bin" \
       bs=1M count=1 status=none
done

echo
echo "Dataset created."

find "$ROOT" -type f | wc -l
du -sh "$ROOT"