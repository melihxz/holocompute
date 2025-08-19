package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/melihxz/holocompute/internal/config"
	"github.com/melihxz/holocompute/internal/dsm"
	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
	"github.com/melihxz/holocompute/internal/membership"
	"github.com/melihxz/holocompute/internal/scheduler"
	"github.com/melihxz/holocompute/pkg/proto"
)

func main() {
	// Create a logger
	logger := log.New(slog.LevelDebug)

	// Create configuration
	cfg := config.DefaultConfig()

	// Create local node info
	localNode := hyperbus.NodeInfo{
		ID:           hyperbus.NodeID(cfg.Node.ID),
		Address:      nil, // This would be a real address in a complete implementation
		PublicKey:    []byte("test-public-key"),
		PQPublicKey:  []byte("test-pq-key"),
		Capabilities: &proto.NodeCapabilities{CpuCores: 4, MemoryBytes: 1024 * 1024 * 1024},
	}

	// Create hyperbus
	bus := hyperbus.New(localNode, nil, logger)

	// Create membership
	localMember := &membership.Member{
		ID:           hyperbus.NodeID(cfg.Node.ID),
		Address:      nil,
		LastSeen:     time.Now(),
		Status:       membership.Alive,
		Capabilities: &proto.NodeCapabilities{CpuCores: 4, MemoryBytes: 1024 * 1024 * 1024},
	}

	_ = membership.NewMembership(localMember, logger)

	// Create memory manager
	memoryMgr := dsm.NewMemoryManager(bus, logger)

	// Create scheduler
	taskScheduler := scheduler.NewScheduler(logger)

	// Start the scheduler
	ctx, cancel := context.WithCancel(context.Background())
	taskScheduler.Start(ctx)
	defer func() {
		cancel()
		taskScheduler.Stop()
	}()

	// Test creating an array
	array, err := memoryMgr.CreateArray(ctx, 1000)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create array:", err)
		os.Exit(1)
	}

	fmt.Printf("Created array with ID: %s and %d pages\n", array.ID, array.PageCount())

	// Test ParallelFor
	results := make([]int, 100)
	fn := func(i int) error {
		results[i] = i * 2
		return nil
	}

	err = scheduler.ParallelFor(ctx, logger, 100, fn, 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ParallelFor failed:", err)
		os.Exit(1)
	}

	fmt.Println("ParallelFor completed successfully")

	// Verify results
	success := true
	for i := 0; i < 100; i++ {
		if results[i] != i*2 {
			success = false
			break
		}
	}

	if success {
		fmt.Println("All results verified correctly")
	} else {
		fmt.Println("Some results are incorrect")
	}

	// Test Map
	in := []int{1, 2, 3, 4, 5}
	out := make([]int, 5)

	mapFn := func(x int) (int, error) {
		return x * x, nil
	}

	err = scheduler.Map(ctx, logger, in, mapFn, out, 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Map failed:", err)
		os.Exit(1)
	}

	fmt.Println("Map completed successfully")
	fmt.Printf("Input: %v\n", in)
	fmt.Printf("Output: %v\n", out)

	// Test Reduce
	var sum int
	reduceFn := func(a, b int) int {
		return a + b
	}

	err = scheduler.Reduce(ctx, logger, in, mapFn, reduceFn, &sum, 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reduce failed:", err)
		os.Exit(1)
	}

	fmt.Printf("Reduce completed successfully. Sum of squares: %d\n", sum)

	fmt.Println("End-to-end test completed successfully!")
}
