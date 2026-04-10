package coordinator

import (
	"log"
	"sync"
	"time"

	"fractal-cluster/internal/compute"
)

type WorkerEntry struct {
	Name     string
	engine   compute.Engine
	close    func()
	lastSeen time.Time
	local    bool
}

type Registry struct {
	mu               sync.RWMutex
	workers          map[string]*WorkerEntry
	names            []string // for round-robin
	nextIdx          int
	heartbeatTimeout time.Duration
	stopCh           chan struct{}
}

func NewRegistry(heartbeatTimeout time.Duration) *Registry {
	return &Registry{
		workers:          make(map[string]*WorkerEntry),
		heartbeatTimeout: heartbeatTimeout,
		stopCh:           make(chan struct{}),
	}
}

// RegisterEngine adds a worker by Engine. The gRPC registration handler
// passes a GRPCEngine; the all-in-one binary passes a LocalEngine with
// local=true so the reaper leaves it alone.
func (r *Registry) RegisterEngine(name string, engine compute.Engine, closeFn func(), local bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, exists := r.workers[name]; exists {
		w.lastSeen = time.Now()
		log.Printf("Worker %s re-registered", name)
		return
	}

	r.workers[name] = &WorkerEntry{
		Name:     name,
		engine:   engine,
		close:    closeFn,
		lastSeen: time.Now(),
		local:    local,
	}
	r.rebuildNames()

	log.Printf("Worker %s registered (total: %d)", name, len(r.workers))
}

func (r *Registry) RecordHeartbeat(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, exists := r.workers[name]; exists {
		w.lastSeen = time.Now()
	}
}

// StartReaper periodically removes workers that missed their heartbeat.
// Local (in-process) workers are exempt.
func (r *Registry) StartReaper(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-r.stopCh:
				return
			case <-ticker.C:
				r.reap()
			}
		}
	}()
}

func (r *Registry) reap() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for name, w := range r.workers {
		if w.local {
			continue
		}
		if now.Sub(w.lastSeen) > r.heartbeatTimeout {
			log.Printf("Worker %s timed out, removing (total: %d)", name, len(r.workers)-1)
			if w.close != nil {
				w.close()
			}
			delete(r.workers, name)
		}
	}
	r.rebuildNames()
}

func (r *Registry) rebuildNames() {
	r.names = make([]string, 0, len(r.workers))
	for name := range r.workers {
		r.names = append(r.names, name)
	}
	if r.nextIdx >= len(r.names) {
		r.nextIdx = 0
	}
}

// NextEngine returns the next worker engine using round-robin.
func (r *Registry) NextEngine() compute.Engine {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(r.names)
	if n == 0 {
		return nil
	}

	name := r.names[r.nextIdx%n]
	r.nextIdx = (r.nextIdx + 1) % n
	return r.workers[name].engine
}

func (r *Registry) HealthyCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.workers)
}

func (r *Registry) Close() {
	close(r.stopCh)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range r.workers {
		if w.close != nil {
			w.close()
		}
	}
}
