# Kaiten

[![Made with Go][made-with-go]](https://go.dev)
[![GoDoc Reference](https://pkg.go.dev/badge/github.com/Gthamsrim1/kaiten)](https://pkg.go.dev/github.com/Gthamsrim1/kaiten)
[![License][license-badge]][license-url]
[![GitHub stars](https://img.shields.io/github/stars/Gthamsrim1/kaiten)](https://github.com/Gthamsrim1/kaiten/stargazers)


[license-badge]: https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square
[license-url]: https://github.com/Gthamsrim1/kaiten/blob/main/LICENSE
[made-with-go]: https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white

Kaiten is an experimental container runtime built in Go that combines a content-addressed snapshot store with a virtual filesystem and Linux namespace isolation. It imports Linux root filesystems, stores them as deduplicated snapshots, and executes them inside isolated runtime environments.

> **Project Status:** Experimental (v0.1.0)

## Features

- Content-addressed object storage
- Content-defined chunking with deduplication
- Snapshot creation and restoration
- FUSE-based virtual filesystem
- Lazy loading of file contents
- Linux namespace isolation (PID, Mount, IPC, UTS)
- `pivot_root()`-based root filesystem switching
- Symbolic link support
- Import existing Linux root filesystems
- Snapshot history, checkout, and garbage collection

## Requirements

- Linux
- Go 1.24+
- FUSE
- Root privileges (required for running containers)

## Building

```bash
git clone https://github.com/Gthamsrim1/kaiten.git
cd kaiten

go build -o kaiten ./cmd/kaiten
```

## Usage

### Import a root filesystem

```bash
./kaiten import /path/to/rootfs
```

Example using Alpine:

```bash
./kaiten import /tmp/alpine/rootfs
```

### List snapshots

```bash
./kaiten snapshots
```

### View snapshot history

```bash
./kaiten log
```

### Checkout a snapshot

```bash
./kaiten checkout <snapshot-id>
```

### Run a container

```bash
sudo ./kaiten run /bin/sh
```

or

```bash
sudo ./kaiten run echo Hello, Kaiten!
```

### Garbage collect unused objects

```bash
./kaiten gc
```

## Project Structure

```
cmd/            CLI
internal/
    chunk/      Content-defined chunking
    persist/    Snapshot persistence
    runtime/    Container runtime
    tree/       Virtual filesystem
```

## Roadmap

- cgroup v2 support
- Container networking
- Bind mounts
- `/dev` management
- User namespaces
- OCI image support

## License

