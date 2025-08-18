# HoloCompute

A distributed memory + compute virtualization layer that turns a set of heterogeneous devices into one logical "personal supercomputer."

## Repository

This repository has moved to [github.com/melihxz/holocompute](https://github.com/melihxz/holocompute).

## Features

- **Unified memory**: Expose a large, cluster-wide virtual memory space
- **Distributed compute**: High-level parallel API with fault tolerance
- **Heterogeneous hardware**: CPU by default, optional GPU offload and WASM sandbox
- **Security by default**: Encrypted transport with PQ-hybrid KEM option
- **Resilience**: Node churn and failures don't crash jobs
- **Observability**: First-class metrics, tracing, and UIs

## Architecture

1. **Hyperbus (network core)** - QUIC transport with Noise encryption
2. **Membership & control plane** - SWIM gossip and Raft consensus
3. **Distributed Shared Memory (DSM)** - Page-granular shared memory
4. **Task Sharder & Scheduler** - Work-stealing scheduler with data locality
5. **Sandbox & plug-ins** - WASM sandbox and GPU plugin API
6. **Storage & persistence** - Badger/Pebble for metadata
7. **Observability** - Prometheus metrics, OpenTelemetry tracing
8. **CLI tooling** - `holo` command-line tool

## Getting Started

```bash
# Build the project
make build

# Run tests
make test

# Run linters
make lint
```

## Documentation

- [Architecture](docs/architecture.md)
- [Consistency Model](docs/consistency.md)
- [Threat Model](docs/threat-model.md)