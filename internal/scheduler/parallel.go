package scheduler

import (
	"context"
	"sync"

	"github.com/melihxz/holocompute/internal/log"
	"golang.org/x/sync/errgroup"
)

// ParallelFor executes a function in parallel for indices 0 to n-1
func ParallelFor(ctx context.Context, logger *log.Logger, n int, fn func(i int) error, maxConcurrency int) error {
	// Create an error group
	g, ctx := errgroup.WithContext(ctx)

	// Set the maximum number of goroutines
	if maxConcurrency > 0 {
		g.SetLimit(maxConcurrency)
	}

	// Submit tasks for each index
	for i := 0; i < n; i++ {
		i := i // Capture loop variable
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return fn(i)
			}
		})
	}

	// Wait for all tasks to complete
	return g.Wait()
}

// Map applies a function to each element of a slice and stores the result in another slice
func Map[T, U any](ctx context.Context, logger *log.Logger, in []T, fn func(T) (U, error), out []U, maxConcurrency int) error {
	if len(in) != len(out) {
		return ErrSliceLengthMismatch
	}

	// Create an error group
	g, ctx := errgroup.WithContext(ctx)

	// Set the maximum number of goroutines
	if maxConcurrency > 0 {
		g.SetLimit(maxConcurrency)
	}

	// Submit tasks for each element
	for i := 0; i < len(in); i++ {
		i := i // Capture loop variable
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				result, err := fn(in[i])
				if err != nil {
					return err
				}
				out[i] = result
				return nil
			}
		})
	}

	// Wait for all tasks to complete
	return g.Wait()
}

// Reduce applies a reduction function to a slice
func Reduce[T, U any](ctx context.Context, logger *log.Logger, in []T, mapFn func(T) (U, error), reduceFn func(U, U) U, result *U, maxConcurrency int) error {
	// First, map all elements
	mapped := make([]U, len(in))
	mapErr := Map(ctx, logger, in, mapFn, mapped, maxConcurrency)
	if mapErr != nil {
		return mapErr
	}

	// Then reduce the mapped elements
	if len(mapped) == 0 {
		var zero U
		*result = zero
		return nil
	}

	// Use a mutex to protect the result
	var mu sync.Mutex
	*result = mapped[0]

	// Create an error group
	g, ctx := errgroup.WithContext(ctx)

	// Set the maximum number of goroutines
	if maxConcurrency > 0 {
		g.SetLimit(maxConcurrency)
	}

	// Submit reduction tasks
	for i := 1; i < len(mapped); i++ {
		i := i // Capture loop variable
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				mu.Lock()
				*result = reduceFn(*result, mapped[i])
				mu.Unlock()
				return nil
			}
		})
	}

	// Wait for all tasks to complete
	return g.Wait()
}

// ErrSliceLengthMismatch is returned when input and output slices have different lengths
var ErrSliceLengthMismatch = &errSliceLengthMismatch{}

type errSliceLengthMismatch struct{}

func (e *errSliceLengthMismatch) Error() string {
	return "input and output slices must have the same length"
}
