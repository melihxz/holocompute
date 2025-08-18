# HoloCompute Architecture

## Overview

HoloCompute is a distributed memory + compute virtualization layer that turns a set of heterogeneous devices into one logical "personal supercomputer." It provides a unified memory space and distributed compute capabilities with strong consistency guarantees, fault tolerance, and security.

## Core Components

### 1. Hyperbus (Network Core)

The Hyperbus is the foundational network layer that provides secure, reliable communication between nodes.

**Key Features:**
- **Transport**: QUIC protocol for connection-oriented, multiplexed streams with congestion control
- **Security**: Noise protocol framework with hybrid post-quantum KEM (X25519 + ML-KEM-768)
- **Message Framing**: Protocol Buffers for efficient serialization
- **Stream Types**: Control Plane (cluster management) and Data Plane (memory/compute operations)

### 2. Membership & Control Plane

The membership layer manages cluster topology and consensus.

**Key Components:**
- **SWIM Gossip**: Scalable Weakly-consistent Infection-style process group Membership for failure detection
- **Raft Consensus**: For cluster state management (ring assignments, shard mappings)
- **Node Inventory**: Tracks hardware capabilities (CPU, memory, GPU, NUMA topology)
- **Placement Algorithms**: Rendezvous hashing for shard ownership, consistent hash rings for resource classes

### 3. Distributed Shared Memory (DSM)

The DSM provides a unified virtual memory space across the cluster.

**Key Features:**
- **Abstractions**: `SharedArray[T]` and `DistBuffer` with page-granular layout
- **Page Management**: 64 KiB pages with locality-aware caching (2Q/ARC)
- **Coherence Model**: Lease-based write ownership (single-writer/multi-reader)
- **Paging Protocol**: Background prefetching, compression (LZ4/Zstd), checksums (BLAKE3)
- **Eviction Policy**: Cost-aware eviction based on size, latency, and access heat

### 4. Task Scheduler

The scheduler orchestrates distributed computation across the cluster.

**Key Features:**
- **API**: `ParallelFor`, `Map`, `Reduce`, `SubmitTask`
- **Scheduling**: Work-stealing with data locality scoring
- **Fault Tolerance**: Task replay, idempotency tokens, speculative execution
- **Flow Control**: Credit-based backpressure, adaptive concurrency

### 5. Sandboxing & Plugins

Secure execution environment for user code and hardware acceleration.

**Key Components:**
- **WASM Sandbox**: Deterministic execution with restricted syscalls
- **GPU Plugins**: gRPC interface for vector/matrix operations
- **Serialization**: CBOR/Protobuf contracts for data exchange

### 6. Storage & Persistence

Persistent storage for cluster metadata and optional page caching.

**Key Features:**
- **Metadata Store**: BadgerDB/Pebble for Raft logs and DSM metadata
- **Page Spill**: Optional encrypted local disk cache
- **Tunables**: Configurable cache sizes and spill thresholds

### 7. Observability

Comprehensive monitoring and debugging capabilities.

**Components:**
- **Metrics**: Prometheus endpoint with key performance indicators
- **Tracing**: OpenTelemetry spans across the request lifecycle
- **Profiling**: pprof endpoints for performance analysis
- **UIs**: TUI ("HoloTop") and Web UI for cluster visualization

### 8. CLI Tooling

Command-line interface for cluster management and operations.

**Commands:**
- `holo agent`: Run a node
- `holo join/leave`: Cluster membership
- `holo status/top`: Cluster monitoring
- `holo alloc`: Resource allocation
- `holo run`: Task execution
- `holo drain`: Node maintenance

## Data Flow

1. **Cluster Formation**: Nodes join using SWIM gossip, establish QUIC connections via Hyperbus
2. **State Consensus**: Raft elects leaders and replicates cluster state (rings, shard maps)
3. **Memory Allocation**: Client allocates `SharedArray`, control plane assigns pages to nodes
4. **Data Access**: Reader caches fetch pages; writers acquire leases from control plane
5. **Computation**: Scheduler places tasks near data, executes via WASM or plugins
6. **Synchronization**: `Sync()` barriers flush writes, invalidate reader caches, bump versions

## Security Model

- **Transport Security**: Noise over QUIC with XChaCha20-Poly1305 or AES-GCM
- **Post-Quantum**: Hybrid KEM (X25519 + ML-KEM-768) with feature flag
- **Identity**: Ed25519 node identities with trust-on-first-use or pre-shared keys
- **Access Control**: RBAC with roles bound to node/user tokens
- **Rate Limiting**: Control endpoint protection
- **Constant-Time**: Key comparisons to prevent timing attacks

## Consistency Model

- **Single Writer**: Lease-based exclusive write access per page
- **Multi Reader**: Versioned immutable pages for concurrent reads
- **Synchronization**: Explicit `Sync()` barriers for deterministic phases
- **Failure Recovery**: Lease expiration on writer crash, Raft log replay for state recovery

## Performance Targets

- **Page Fetch**: P50 < 2ms, P99 < 10ms for 64 KiB over 1 Gbps LAN
- **Speedup**: 3.2× on 4-node `ParallelFor` of 100M elements vs. single node
- **Overhead**: < 5% CPU on idle cluster
- **Backpressure**: Bounded queues, adaptive concurrency control