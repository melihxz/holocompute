package scheduler

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestScheduler_SubmitTask(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create scheduler
	scheduler := NewScheduler(logger)

	// Start the scheduler
	ctx, cancel := context.WithCancel(context.Background())
	scheduler.Start(ctx)

	// Create a task
	task := &Task{
		ID:       "test-task",
		Function: func() error { time.Sleep(time.Millisecond * 10); return nil },
		Result:   make(chan error, 1),
	}

	// Submit the task
	err := scheduler.SubmitTask(ctx, task)
	assert.NoError(t, err)

	// Wait for the result
	select {
	case result := <-task.Result:
		assert.NoError(t, result)
	case <-time.After(time.Second):
		t.Fatal("task did not complete within timeout")
	}

	// Stop the scheduler
	cancel()
	scheduler.Stop()
}

func TestParallelFor(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	ctx := context.Background()

	// Test ParallelFor with a simple function
	results := make([]int, 100)
	fn := func(i int) error {
		results[i] = i * 2
		return nil
	}

	err := ParallelFor(ctx, logger, 100, fn, 10)
	assert.NoError(t, err)

	// Verify results
	for i := 0; i < 100; i++ {
		assert.Equal(t, i*2, results[i])
	}
}

func TestMap(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	ctx := context.Background()

	// Test Map with a simple function
	in := []int{1, 2, 3, 4, 5}
	out := make([]int, 5)

	fn := func(x int) (int, error) {
		return x * x, nil
	}

	err := Map(ctx, logger, in, fn, out, 5)
	assert.NoError(t, err)

	// Verify results
	expected := []int{1, 4, 9, 16, 25}
	assert.Equal(t, expected, out)
}

func TestReduce(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	ctx := context.Background()

	// Test Reduce with a simple function
	in := []int{1, 2, 3, 4, 5}
	var result int

	mapFn := func(x int) (int, error) {
		return x, nil
	}

	reduceFn := func(a, b int) int {
		return a + b
	}

	err := Reduce(ctx, logger, in, mapFn, reduceFn, &result, 5)
	assert.NoError(t, err)

	// Verify result
	assert.Equal(t, 15, result) // 1+2+3+4+5 = 15
}
