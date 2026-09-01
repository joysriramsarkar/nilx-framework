// Package async provides concurrency primitives, timers, and futures for NilLang.
package async

import (
	"sync"
	"time"
)

// Task represents a background asynchronous unit of execution.
type Task struct {
	done chan struct{}
	err  error
	res  interface{}
}

// Spawn runs a function in a concurrent goroutine.
func Spawn(fn func() (interface{}, error)) *Task {
	t := &Task{done: make(chan struct{})}
	go func() {
		defer close(t.done)
		res, err := fn()
		t.res = res
		t.err = err
	}()
	return t
}

// Await waits for the task to complete and returns its result.
func (t *Task) Await() (interface{}, error) {
	<-t.done
	return t.res, t.err
}

// Delay returns a task that completes after ms milliseconds.
func Delay(ms int64) *Task {
	return Spawn(func() (interface{}, error) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return nil, nil
	})
}

// Parallel runs multiple tasks concurrently and waits for all of them.
func Parallel(tasks ...func() (interface{}, error)) []interface{} {
	var wg sync.WaitGroup
	results := make([]interface{}, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, fn func() (interface{}, error)) {
			defer wg.Done()
			res, _ := fn()
			results[idx] = res
		}(i, task)
	}

	wg.Wait()
	return results
}
