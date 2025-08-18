# HoloCompute Consistency Model

## Overview

HoloCompute provides a carefully designed consistency model that balances performance with correctness. It uses a lease-based coherence protocol with explicit synchronization barriers to provide predictable behavior for distributed computations.

## Core Principles

### 1. Single Writer, Multiple Readers (SWMR)

Each page in a `SharedArray` follows a Single Writer, Multiple Readers (SWMR) model:
- Only one node can hold a write lease for a page at any time
- Multiple nodes can hold read leases simultaneously
- Readers access versioned, immutable snapshots of pages
- Writers modify pages exclusively and atomically

### 2. Lease-Based Coherence

Page access is controlled through leases managed by the control plane:
- **Read Leases**: Allow reading a specific version of a page
- **Write Leases**: Allow reading and modifying a page
- **Lease Duration**: Short-lived (default: seconds) to minimize blocking
- **Lease Expiration**: Automatic cleanup of stale leases on node failure

### 3. Explicit Synchronization

Deterministic computation phases are achieved through explicit synchronization:
- `Sync()` operations act as global barriers
- All pending writes are flushed before barrier completion
- Reader caches are invalidated after barrier completion
- Array versions are incremented to ensure visibility of new data

## Detailed Behavior

### Read Operations

1. **Cache Check**: Reader first checks local page cache
2. **Lease Acquisition**: If not cached, acquire read lease from control plane
3. **Page Fetch**: Retrieve page from owner node if needed
4. **Local Caching**: Cache page locally for subsequent accesses
5. **Data Access**: Access data from cached page

### Write Operations

1. **Lease Check**: Verify no conflicting leases exist
2. **Lease Acquisition**: Acquire write lease from control plane
3. **Page Fetch**: Retrieve current page content from owner
4. **Local Modification**: Modify page in local memory
5. **Write Buffering**: Queue page for eventual flush

### Synchronization

1. **Write Flush**: All buffered writes are flushed to owner nodes
2. **Lease Revocation**: All write leases are revoked
3. **Cache Invalidation**: Reader caches are invalidated cluster-wide
4. **Version Bump**: Array version is incremented
5. **Barrier Completion**: Sync operation returns to caller

## Failure Handling

### Writer Failure

If a writer node fails before committing:
1. **Lease Expiration**: Write lease expires after timeout
2. **State Reversion**: Page reverts to previous version
3. **Task Replay**: Failed tasks are rescheduled elsewhere
4. **No Data Loss**: Committed data remains consistent

### Reader Failure

If a reader node fails:
1. **Lease Cleanup**: Read leases are automatically cleaned up
2. **No Impact**: No effect on writers or other readers
3. **Cache Loss**: Only local cache is lost (re-fetch on next access)

### Network Partition

During network partitions:
1. **Local Continuation**: Nodes continue with cached data
2. **Lease Expiration**: Remote leases expire in partition
3. **Partitioned Operation**: Each partition operates independently
4. **Merge Recovery**: Partitions reconcile on reconnection

## Guarantees

### Strong Consistency

- **Atomic Writes**: Page modifications are atomic
- **Sequential Consistency**: Sync() provides sequential consistency
- **Linearizability**: Within a single page, operations appear linearizable
- **Durability**: Committed writes survive node failures

### Eventual Consistency

- **Cache Coherency**: Reader caches eventually reflect latest writes
- **Membership Updates**: Cluster topology changes propagate eventually
- **Metric Aggregation**: Observability data is eventually consistent

## Programming Model

### Deterministic Parallelism

```go
// All writes visible to subsequent reads after Sync()
err := c.ParallelFor(1000, func(i int) error {
    return arr.Set(i, compute(i))  // Writer 1
})
arr.Sync()  // Barrier - all writes committed

err = c.ParallelFor(1000, func(i int) error {
    v, _ := arr.Get(i)  // Reader sees committed values
    return arr.Set(i, v*2)  // Writer 2 (different pages OK)
})
```

### Conflict Detection

```go
// Writer conflict - second writer blocked until first commits
go writer1() // Acquires write lease on page 0
go writer2() // Blocks waiting for page 0 lease

func writer1() {
    arr.Set(0, 42)  // Acquires lease
    time.Sleep(time.Second)
    arr.Sync()      // Releases lease
}

func writer2() {
    arr.Set(0, 84)  // Waits for lease, then succeeds
}
```

## Performance Considerations

### Lease Overhead

- **Acquisition**: Single RTT to control plane
- **Renewal**: Background renewal before expiration
- **Revocation**: Async notification to lease holders

### Cache Efficiency

- **Hit Ratio**: High for repeated accesses
- **Prefetching**: Background fetch of adjacent pages
- **Eviction**: Cost-aware policy minimizes re-fetches

### Synchronization Cost

- **Latency**: Proportional to pending write volume
- **Throughput**: Pipeline-friendly with overlapping barriers
- **Scalability**: Distributed coordination with minimal contention