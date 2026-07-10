#!/usr/bin/env bash
set -e

rm -rf kaiten-data

hyperfine \
    --warmup 2 \
    './kaiten import /tmp/kaiten-bench'

echo
echo "Stored objects:"
find kaiten-data/objects -type f | wc -l

echo
echo "Repository size:"
du -sh kaiten-data