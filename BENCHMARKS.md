# Benchmarks

All benchmarks were collected on a development build of Kaiten.

## Test Environment

| Component | Value |
|----------|-------|
| Language | Go 1.26 |
| Operating System | Kali Linux |
| Filesystem | ext4 |
| Benchmark Tool | hyperfine, GNU time |

---

# Import Performance

## Alpine Linux Root Filesystem

| Metric | Value |
|--------|------:|
| Files | 84 |
| Directories | 97 |
| Symbolic Links | 335 |
| Total Filesystem Objects | **516** |
| Dataset Size | **8.7 MB** |
| Content Objects Stored | **177** |
| Import Time | **270 ms** |

---

# Deduplication

Re-importing the same Alpine root filesystem produced:

| Metric | Value |
|--------|------:|
| Additional Content Objects | **0** |

This demonstrates complete content deduplication through content-addressed storage.

---

# Lazy Loading

Synthetic benchmark using a 5 GB filesystem containing **5,000 × 1 MB** random files.

| Metric | Lazy Loading | Eager Loading |
|--------|-------------:|--------------:|
| Data Loaded During Restore | On Demand | 5.0 GB |
| Files Loaded During Restore | On Demand | 5,000 |
| Runtime Startup | **390 ms** | **4.62 s** |

The benchmark measures runtime initialization before executing the target application. Lazy loading avoids reading file contents until they are first accessed, significantly reducing startup latency for large snapshots.

---

# Benchmark Methodology

## Import

```bash
time ./kaiten import <rootfs>
```

## Startup

```bash
time sudo ./kaiten run /bin/true
```

## Deduplication

```bash
./kaiten import <rootfs>
./kaiten import <rootfs>

find kaiten-data/objects -type f | wc -l
```