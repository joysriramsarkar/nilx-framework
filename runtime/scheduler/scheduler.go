// Package scheduler implements the cooperative task scheduler for NilRT.
package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents an asynchronous green thread managed by NilRT.
type Task struct {
	ID        uint64
	Name      string
	Work      func() error
	completed chan struct{}
	err       error
}

// Scheduler multiplexes NilLang tasks across worker threads.
type Scheduler struct {
	mu         sync.RWMutex
	taskQueue  chan *Task
	workerCnt  int
	nextID     uint64
	ctx        context.Context
	cancel     context.CancelFunc
	activeTask uint64
}

// New creates a new task scheduler with worker pool.
func New(workers int) *Scheduler {
	if workers <= 0 {
		workers = 4
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		taskQueue: make(chan *Task, 128),
		workerCnt: workers,
		ctx:       ctx,
		cancel:    cancel,
	}
	s.startWorkers()
	return s
}

func (s *Scheduler) startWorkers() {
	for i := 0; i < s.workerCnt; i++ {
		go func() {
			for {
				select {
				case <-s.ctx.Done():
					return
				case t, ok := <-s.taskQueue:
					if !ok {
						return
					}
					atomic.AddUint64(&s.activeTask, 1)
					t.err = t.Work()
					atomic.AddUint64(&s.activeTask, ^uint64(0))
					close(t.completed)
				}
			}
		}()
	}
}

// Spawn schedules a new asynchronous task.
func (s *Scheduler) Spawn(name string, work func() error) *Task {
	id := atomic.AddUint64(&s.nextID, 1)
	t := &Task{
		ID:        id,
		Name:      name,
		Work:      work,
		completed: make(chan struct{}),
	}
	s.taskQueue <- t
	return t
}

// Await blocks until the given task completes.
func (t *Task) Await() error {
	<-t.completed
	return t.err
}

// AwaitTimeout blocks until the task completes or timeout expires.
func (t *Task) AwaitTimeout(d time.Duration) error {
	select {
	case <-t.completed:
		return t.err
	case <-time.After(d):
		return context.DeadlineExceeded
	}
}

// Stop terminates the scheduler workers.
func (s *Scheduler) Stop() {
	s.cancel()
}
