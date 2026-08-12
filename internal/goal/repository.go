package goal

import (
	"sync"
	"time"
)

type Repository struct {
	mu      sync.RWMutex
	targets map[int]int
}

func NewRepository() *Repository {
	repo := &Repository{
		targets: make(map[int]int),
	}
	repo.targets[time.Now().Year()] = 24
	return repo
}

func (r *Repository) GetTarget(year int) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	target, ok := r.targets[year]
	if !ok {
		return 0
	}
	return target
}

func (r *Repository) SetTarget(year int, target int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.targets[year] = target
}
