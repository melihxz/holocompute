package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/melihxz/holocompute/pkg/holocompute"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Connect to the cluster
	c, err := holocompute.Connect(ctx, holocompute.Options{
		Bootstrap: []string{"127.0.0.1:8443"},
	})
	if err != nil {
		log.Fatal("Failed to connect to cluster:", err)
	}

	// Create input arrays
	fmt.Println("Creating input arrays...")
	start := time.Now()
	arrA, err := c.NewSharedArray(10_000_000, holocompute.Policy{
		Replication: 1,
	})
	if err != nil {
		log.Fatal("Failed to create array A:", err)
	}

	arrB, err := c.NewSharedArray(10_000_000, holocompute.Policy{
		Replication: 1,
	})
	if err != nil {
		log.Fatal("Failed to create array B:", err)
	}

	// Fill input arrays
	fmt.Println("Filling input arrays...")
	err = c.ParallelFor(arrA.Len(), func(i int) error {
		return arrA.Set(i, float32(i)*0.5)
	})
	if err != nil {
		log.Fatal("Failed to fill array A:", err)
	}

	err = c.ParallelFor(arrB.Len(), func(i int) error {
		return arrB.Set(i, float32(i)*0.3)
	})
	if err != nil {
		log.Fatal("Failed to fill array B:", err)
	}

	// Synchronize arrays
	err = arrA.Sync()
	if err != nil {
		log.Fatal("Failed to sync array A:", err)
	}

	err = arrB.Sync()
	if err != nil {
		log.Fatal("Failed to sync array B:", err)
	}

	fmt.Printf("Input arrays created and filled in %v\n", time.Since(start))

	// Create output array
	arrC, err := c.NewSharedArray(10_000_000, holocompute.Policy{
		Replication: 1,
	})
	if err != nil {
		log.Fatal("Failed to create array C:", err)
	}

	// Load WASM module
	fmt.Println("Loading WASM module...")
	mod := holocompute.MustLoadWASM("kernels/vector_add.wasm")

	// Create task specification
	task := holocompute.TaskSpec{
		Module: mod,
		Func:   "vec_add",
		Inputs: holocompute.Inputs{
			"A": arrA,
			"B": arrB,
		},
		Outputs: holocompute.Outputs{
			"C": arrC,
		},
		ResourceHints: holocompute.ResourceHints{
			CPU:      2,
			GPU:      false,
			MemoryMB: 100,
		},
	}

	// Submit task
	fmt.Println("Submitting vector addition task...")
	start = time.Now()
	res, err := c.SubmitTask(ctx, task)
	if err != nil {
		log.Fatal("Failed to submit task:", err)
	}

	fmt.Printf("Task completed in %v\n", time.Since(start))
	fmt.Printf("Task status: %v\n", res.Status)

	// Verify a few results
	fmt.Println("Verifying results...")
	for i := 0; i < 10; i++ {
		a, _ := arrA.Get(i)
		b, _ := arrB.Get(i)
		c, _ := arrC.Get(i)
		// In a real implementation, we would do proper type assertions
		fmt.Printf("Index %d: A=%v, B=%v, C=%v\n", i, a, b, c)
	}

	// Clean up
	arrA.Close()
	arrB.Close()
	arrC.Close()
}
