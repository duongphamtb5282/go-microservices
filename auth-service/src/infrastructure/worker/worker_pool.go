package worker

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"backend-core/logging"
)

// Task represents a unit of work to be executed by a worker
type Task struct {
	ID         string
	Type       string
	Payload    interface{}
	CreatedAt  time.Time
	Priority   int // 1=low, 2=normal, 3=high, 4=critical
	RetryCount int
	MaxRetries int
	Timeout    time.Duration
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID    string
	Success   bool
	Error     error
	Duration  time.Duration
	StartedAt time.Time
	EndedAt   time.Time
}

// TaskHandler defines the interface for handling different task types
type TaskHandler interface {
	HandleTask(ctx context.Context, task *Task) error
	GetTaskType() string
}

// WorkerPool manages a pool of goroutines for concurrent task processing
type WorkerPool struct {
	name       string
	numWorkers int
	taskQueue  chan *Task
	resultChan chan *TaskResult
	handlers   map[string]TaskHandler
	logger     *logging.Logger

	// Control channels
	stopChan chan struct{}
	stopOnce sync.Once

	// Metrics
	tasksProcessed int64
	tasksFailed    int64
	activeWorkers  int64
	queueSize      int64

	// Synchronization
	wg sync.WaitGroup
	mu sync.RWMutex
}

// WorkerPoolConfig holds configuration for the worker pool
type WorkerPoolConfig struct {
	Name           string
	NumWorkers     int
	QueueSize      int
	MaxRetries     int
	DefaultTimeout time.Duration
}

// NewWorkerPool creates a new worker pool with the given configuration
func NewWorkerPool(config *WorkerPoolConfig, logger *logging.Logger) *WorkerPool {
	if config.NumWorkers <= 0 {
		config.NumWorkers = runtime.GOMAXPROCS(0) * 2
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 1000
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Second
	}

	wp := &WorkerPool{
		name:       config.Name,
		numWorkers: config.NumWorkers,
		taskQueue:  make(chan *Task, config.QueueSize),
		resultChan: make(chan *TaskResult, config.QueueSize),
		handlers:   make(map[string]TaskHandler),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}

	// Register default handlers
	wp.registerDefaultHandlers()

	// Start workers
	wp.startWorkers()

	logger.Info("Worker pool started",
		logging.String("pool_name", config.Name),
		logging.Int("num_workers", config.NumWorkers),
		logging.Int("queue_size", config.QueueSize))

	return wp
}

// RegisterHandler registers a task handler for a specific task type
func (wp *WorkerPool) RegisterHandler(handler TaskHandler) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.handlers[handler.GetTaskType()] = handler
	wp.logger.Info("Registered task handler",
		logging.String("pool_name", wp.name),
		logging.String("task_type", handler.GetTaskType()))
}

// SubmitTask submits a task to the worker pool
func (wp *WorkerPool) SubmitTask(ctx context.Context, task *Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("%s-%d", task.Type, time.Now().UnixNano())
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}
	if task.Timeout == 0 {
		task.Timeout = 30 * time.Second
	}

	select {
	case wp.taskQueue <- task:
		wp.mu.Lock()
		wp.queueSize++
		wp.mu.Unlock()

		wp.logger.Debug("Task submitted to worker pool",
			logging.String("pool_name", wp.name),
			logging.String("task_id", task.ID),
			logging.String("task_type", task.Type),
			logging.Int("queue_size", wp.GetQueueSize()))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to submit task: %w", ctx.Err())
	case <-wp.stopChan:
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitTaskAsync submits a task asynchronously without waiting
func (wp *WorkerPool) SubmitTaskAsync(task *Task) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := wp.SubmitTask(ctx, task); err != nil {
			wp.logger.Error("Failed to submit task asynchronously",
				logging.Error(err),
				logging.String("task_id", task.ID),
				logging.String("task_type", task.Type))
		}
	}()
}

// GetResultChan returns the channel for receiving task results
func (wp *WorkerPool) GetResultChan() <-chan *TaskResult {
	return wp.resultChan
}

// GetStats returns current pool statistics
func (wp *WorkerPool) GetStats() map[string]interface{} {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return map[string]interface{}{
		"pool_name":       wp.name,
		"num_workers":     wp.numWorkers,
		"active_workers":  wp.activeWorkers,
		"queue_size":      wp.queueSize,
		"tasks_processed": wp.tasksProcessed,
		"tasks_failed":    wp.tasksFailed,
		"queue_capacity":  cap(wp.taskQueue),
		"result_capacity": cap(wp.resultChan),
	}
}

// GetQueueSize returns the current queue size
func (wp *WorkerPool) GetQueueSize() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return int(wp.queueSize)
}

// GetActiveWorkers returns the number of currently active workers
func (wp *WorkerPool) GetActiveWorkers() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return int(wp.activeWorkers)
}

// startWorkers starts the worker goroutines
func (wp *WorkerPool) startWorkers() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is the main worker goroutine
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started",
		logging.String("pool_name", wp.name),
		logging.Int("worker_id", workerID))

	for {
		select {
		case task := <-wp.taskQueue:
			wp.mu.Lock()
			wp.activeWorkers++
			wp.queueSize--
			wp.mu.Unlock()

			wp.processTask(task, workerID)

			wp.mu.Lock()
			wp.activeWorkers--
			wp.mu.Unlock()

		case <-wp.stopChan:
			wp.logger.Debug("Worker stopping",
				logging.String("pool_name", wp.name),
				logging.Int("worker_id", workerID))
			return
		}
	}
}

// processTask processes a single task
func (wp *WorkerPool) processTask(task *Task, workerID int) {
	startTime := time.Now()

	// Create task context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()

	// Find handler for task type
	wp.mu.RLock()
	handler, exists := wp.handlers[task.Type]
	wp.mu.RUnlock()

	if !exists {
		wp.sendResult(&TaskResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Errorf("no handler registered for task type: %s", task.Type),
			Duration:  time.Since(startTime),
			StartedAt: startTime,
			EndedAt:   time.Now(),
		})
		return
	}

	// Execute task with retry logic
	var lastErr error
	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				lastErr = ctx.Err()
				break
			}
		}

		err := handler.HandleTask(ctx, task)
		if err == nil {
			// Success
			wp.mu.Lock()
			wp.tasksProcessed++
			wp.mu.Unlock()

			wp.sendResult(&TaskResult{
				TaskID:    task.ID,
				Success:   true,
				Error:     nil,
				Duration:  time.Since(startTime),
				StartedAt: startTime,
				EndedAt:   time.Now(),
			})
			return
		}

		lastErr = err
		wp.logger.Warn("Task execution failed, retrying",
			logging.String("task_id", task.ID),
			logging.String("task_type", task.Type),
			logging.Int("attempt", attempt+1),
			logging.Int("max_retries", task.MaxRetries),
			logging.Error(err))
	}

	// All retries failed
	wp.mu.Lock()
	wp.tasksFailed++
	wp.mu.Unlock()

	wp.sendResult(&TaskResult{
		TaskID:    task.ID,
		Success:   false,
		Error:     lastErr,
		Duration:  time.Since(startTime),
		StartedAt: startTime,
		EndedAt:   time.Now(),
	})
}

// sendResult sends task result to result channel
func (wp *WorkerPool) sendResult(result *TaskResult) {
	select {
	case wp.resultChan <- result:
		// Result sent successfully
	default:
		// Result channel is full, log warning
		wp.logger.Warn("Result channel is full, dropping result",
			logging.String("task_id", result.TaskID),
			logging.Bool("success", result.Success))
	}
}

// registerDefaultHandlers registers built-in task handlers
func (wp *WorkerPool) registerDefaultHandlers() {
	// Register handlers will be added by service-specific implementations
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	wp.stopOnce.Do(func() {
		close(wp.stopChan)
	})

	// Wait for workers to finish or context timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("Worker pool shutdown completed",
			logging.String("pool_name", wp.name))
		return nil
	case <-ctx.Done():
		wp.logger.Warn("Worker pool shutdown timed out",
			logging.String("pool_name", wp.name))
		return ctx.Err()
	}
}
