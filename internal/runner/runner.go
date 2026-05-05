package runner

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a unit of work
type Task func(ctx context.Context) error

// Pool manages a pool of workers
type Pool struct {
	workers   int
	queueSize int
	tasks     chan Task
	wg        sync.WaitGroup
	errors    chan error
	stopOnce  sync.Once
	stopped   int32
}

// NewPool creates a new worker pool
func NewPool(workers, queueSize int) *Pool {
	if workers <= 0 {
		workers = 10
	}
	if queueSize <= 0 {
		queueSize = workers * 2
	}
	return &Pool{
		workers:   workers,
		queueSize: queueSize,
		tasks:     make(chan Task, queueSize),
		errors:    make(chan error, queueSize),
	}
}

// Start begins processing tasks
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					if err := task(ctx); err != nil {
						select {
						case p.errors <- err:
						default:
						}
					}
				}
			}
		}()
	}
}

// Submit adds a task to the queue (blocking)
func (p *Pool) Submit(task Task) error {
	if atomic.LoadInt32(&p.stopped) == 1 {
		return fmt.Errorf("pool is stopped")
	}
	p.tasks <- task
	return nil
}

// SubmitNonBlocking adds a task without blocking
func (p *Pool) SubmitNonBlocking(task Task) bool {
	if atomic.LoadInt32(&p.stopped) == 1 {
		return false
	}
	select {
	case p.tasks <- task:
		return true
	default:
		return false
	}
}

// Stop gracefully shuts down the pool
func (p *Pool) Stop() {
	p.stopOnce.Do(func() {
		atomic.StoreInt32(&p.stopped, 1)
		close(p.tasks)
	})
}

// Wait blocks until all workers finish
func (p *Pool) Wait() {
	p.wg.Wait()
	close(p.errors)
}

// Errors returns the error channel
func (p *Pool) Errors() <-chan error {
	return p.errors
}

// Stats holds pool statistics
type Stats struct {
	Completed int64
	Failed    int64
	Active    int64
	Queued    int64
}

// Progress tracks scan progress
type Progress struct {
	Total     int64
	Completed int64
	Open      int64
	Closed    int64
	Filtered  int64
	StartTime time.Time
}

// NewProgress creates a new progress tracker
func NewProgress(total int64) *Progress {
	return &Progress{
		Total:     total,
		StartTime: time.Now(),
	}
}

// IncrementOpen increments open port count
func (p *Progress) IncrementOpen() {
	atomic.AddInt64(&p.Completed, 1)
	atomic.AddInt64(&p.Open, 1)
}

// IncrementClosed increments closed port count
func (p *Progress) IncrementClosed() {
	atomic.AddInt64(&p.Completed, 1)
	atomic.AddInt64(&p.Closed, 1)
}

// IncrementFiltered increments filtered port count
func (p *Progress) IncrementFiltered() {
	atomic.AddInt64(&p.Completed, 1)
	atomic.AddInt64(&p.Filtered, 1)
}

// Percentage returns completion percentage
func (p *Progress) Percentage() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(atomic.LoadInt64(&p.Completed)) / float64(p.Total) * 100
}

// ETA returns estimated time of completion
func (p *Progress) ETA() time.Duration {
	completed := atomic.LoadInt64(&p.Completed)
	if completed == 0 {
		return 0
	}
	elapsed := time.Since(p.StartTime)
	rate := float64(completed) / elapsed.Seconds()
	remaining := float64(p.Total-completed) / rate
	return time.Duration(remaining) * time.Second
}

// Rate returns current scan rate (items per second)
func (p *Progress) Rate() float64 {
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(atomic.LoadInt64(&p.Completed)) / elapsed
}

// String returns a formatted progress string
func (p *Progress) String() string {
	return fmt.Sprintf("Progress: %.1f%% (%d/%d) | Open: %d | Rate: %.0f/s | ETA: %s",
		p.Percentage(),
		atomic.LoadInt64(&p.Completed),
		p.Total,
		atomic.LoadInt64(&p.Open),
		p.Rate(),
		p.ETA().Round(time.Second),
	)
}
