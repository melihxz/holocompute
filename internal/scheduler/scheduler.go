package scheduler

import (
	"context"
	"sync"

	"github.com/melihxz/holocompute/internal/log"
)

// Task represents a unit of work to be executed
type Task struct {
	ID       string
	Function func() error
	Result   chan error
	Cancel   context.CancelFunc
}

// Scheduler manages task execution
type Scheduler struct {
	tasks    map[string]*Task
	taskChan chan *Task
	logger   *log.Logger
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewScheduler creates a new task scheduler
func NewScheduler(logger *log.Logger) *Scheduler {
	return &Scheduler{
		tasks:    make(map[string]*Task),
		taskChan: make(chan *Task, 100),
		logger:   logger,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.run(ctx)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.taskChan)
	s.wg.Wait()
}

// SubmitTask submits a task for execution
func (s *Scheduler) SubmitTask(ctx context.Context, task *Task) error {
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	select {
	case s.taskChan <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// run executes tasks from the task channel
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskChan:
			if task == nil {
				return // Channel closed
			}

			// Execute the task in a goroutine
			go s.executeTask(task)

		case <-ctx.Done():
			return
		}
	}
}

// executeTask executes a single task
func (s *Scheduler) executeTask(task *Task) {
	s.logger.Debug("executing task", "task_id", task.ID)

	// Execute the task function
	err := task.Function()

	// Send the result
	select {
	case task.Result <- err:
	default:
		// Result channel is full or closed
		s.logger.Warn("task result channel is full or closed", "task_id", task.ID)
	}

	// Remove the task from the map
	s.mu.Lock()
	delete(s.tasks, task.ID)
	s.mu.Unlock()

	s.logger.Debug("task completed", "task_id", task.ID, "error", err)
}
