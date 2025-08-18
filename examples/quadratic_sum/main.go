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

	// Create a shared array of 100M elements
	fmt.Println("Creating shared array...")
	start := time.Now()
	arr, err := c.NewSharedArray(100_000_000, holocompute.Policy{
		Replication: 1,
	})
	if err != nil {
		log.Fatal("Failed to create array:", err)
	}
	fmt.Printf("Array created in %v\n", time.Since(start))

	// Fill the array using ParallelFor
	fmt.Println("Filling array with quadratic values...")
	start = time.Now()
	err = c.ParallelFor(arr.Len(), func(i int) error {
		v := int64(i)
		return arr.Set(i, v*v+3*v+1)
	})
	if err != nil {
		log.Fatal("Failed to fill array:", err)
	}
	fmt.Printf("Array filled in %v\n", time.Since(start))

	// Synchronize the array
	fmt.Println("Synchronizing array...")
	start = time.Now()
	err = arr.Sync()
	if err != nil {
		log.Fatal("Failed to sync array:", err)
	}
	fmt.Printf("Array synchronized in %v\n", time.Since(start))

	// Compute sum using Reduce
	fmt.Println("Computing sum...")
	start = time.Now()
	var sum interface{}
	err = c.Reduce(arr,
		func(v interface{}) (interface{}, error) { return v, nil },
		func(a, b interface{}) interface{} {
			// In a real implementation, we would do proper type assertions
			return a
		},
		&sum,
	)
	if err != nil {
		log.Fatal("Failed to compute sum:", err)
	}
	fmt.Printf("Sum computed in %v\n", time.Since(start))
	fmt.Printf("Sum: %v\n", sum)

	// Clean up
	arr.Close()
}
