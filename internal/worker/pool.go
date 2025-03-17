package worker

import (
	"sync"
)

// Pool provides a way to run tasks with a controlled number of workers.
type Pool struct {
	wg  sync.WaitGroup
	sem chan struct{}
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
// If workers=1, tasks execute sequentially (one at a time).
func NewWorkerPool(workers int) *Pool {
	if workers <= 0 {
		workers = 1 // Ensure at least one worker
	}

	return &Pool{
		sem: make(chan struct{}, workers),
	}
}

// Submit adds a task to the pool.
func (p *Pool) Submit(task func()) {
	p.wg.Add(1)

	go func() {
		// Acquire worker slot (blocks if all workers are busy)
		p.sem <- struct{}{}

		defer p.wg.Done()
		defer func() { <-p.sem }() // Release worker slot

		task()
	}()
}

// Wait waits for all submitted tasks to complete.
func (p *Pool) Wait() {
	p.wg.Wait()
}
