# FUTURE.md

> Research ideas for Kaiten Runtime

## Fundamentals (can't eliminate)

- Process creation (`clone`, `execve`)
- Namespaces
- cgroups
- CPU execution
- Memory page faults

These are already extremely fast.

---

## Opportunities

### Object Store
Store content-addressed objects instead of image tarballs.
> Eliminates duplicate storage and enables incremental downloads.

### Virtual Filesystem
Never extract images; materialize files only when accessed.
> Reduces startup time and disk usage.

### Page Streaming
Fetch memory pages on page faults instead of entire files.
> Minimizes network transfer for large images.

### Predictive Prefetch
Learn access patterns and fetch data before it's needed.
> Hides network latency.

### Global Chunk Cache
Deduplicate chunks across every container on the machine.
> Saves disk, RAM and bandwidth.

### Build Graphs
Represent images as dependency graphs instead of immutable layers.
> Greatly reduces rebuild time.

### Recipe-based Files
Store deterministic recipes instead of generated files where possible.
> Saves storage for configs and generated assets.

### Execution-oriented Containers
Represent only what a process needs instead of a full filesystem.
> Long-term research direction.

---

## Philosophy

- Never download unused bytes.
- Never store duplicate bytes.
- Never extract unnecessary bytes.
- Generate deterministic data instead of storing it.
- Optimize I/O, not `clone()`.