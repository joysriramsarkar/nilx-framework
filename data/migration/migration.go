package migration

import "sync"

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

type Runner struct {
	migrations []Migration
	applied    map[int]bool
	mu         sync.Mutex
}

func NewRunner() *Runner {
	return &Runner{
		migrations: make([]Migration, 0),
		applied:    make(map[int]bool),
	}
}

func (r *Runner) Register(m Migration) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.migrations = append(r.migrations, m)
	return r
}

func (r *Runner) Up() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	applied := make([]int, 0)
	for _, m := range r.migrations {
		if !r.applied[m.Version] {
			r.applied[m.Version] = true
			applied = append(applied, m.Version)
		}
	}
	return applied
}
