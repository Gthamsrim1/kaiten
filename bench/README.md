# Kaiten Benchmarks

This directory contains reproducible benchmarks for Kaiten.

## Requirements

- hyperfine
- GNU time (`/usr/bin/time`)
- Go build of `kaiten`

Install:

```bash
sudo apt install hyperfine time
```

## Generate a Dataset

Creates a synthetic 5 GB filesystem consisting of 5,000 random 1 MB files.

```bash
./generate_dataset.sh
```

## Import Benchmark

```bash
./benchmark_import.sh
```

Measures:

- Import time
- Repository size
- Number of stored objects

## Startup Benchmark

Benchmark lazy loading:

```bash
./benchmark_startup.sh lazy
```

Benchmark eager loading:

```bash
./benchmark_startup.sh eager
```

Measures:

- Startup latency
- Peak memory usage

## Deduplication Benchmark

```bash
./benchmark_dedup.sh
```

Measures:

- Objects after first import
- Objects after second import
- Additional objects created