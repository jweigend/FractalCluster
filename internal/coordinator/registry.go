package coordinator

import (
	"log"
	"sync"
	"time"

	pb "fractal-cluster/internal/gen/fractal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkerEntry struct {
	Address  string
	conn     *grpc.ClientConn
	client   pb.FractalWorkerClient
	lastSeen time.Time
}

type Registry struct {
	mu               sync.RWMutex
	workers          map[string]*WorkerEntry
	addrs            []string // for round-robin
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

// Register adds a new worker or refreshes an existing one.
func (r *Registry) Register(address string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, exists := r.workers[address]; exists {
		w.lastSeen = time.Now()
		log.Printf("Worker %s re-registered", address)
		return nil
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	r.workers[address] = &WorkerEntry{
		Address:  address,
		conn:     conn,
		client:   pb.NewFractalWorkerClient(conn),
		lastSeen: time.Now(),
	}
	r.rebuildAddrs()

	log.Printf("Worker %s registered (total: %d)", address, len(r.workers))
	return nil
}

// RecordHeartbeat updates the last-seen time for a worker.
func (r *Registry) RecordHeartbeat(address string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, exists := r.workers[address]; exists {
		w.lastSeen = time.Now()
	}
}

// StartReaper periodically removes workers that missed their heartbeat.
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
	for addr, w := range r.workers {
		if now.Sub(w.lastSeen) > r.heartbeatTimeout {
			log.Printf("Worker %s timed out, removing (total: %d)", addr, len(r.workers)-1)
			w.conn.Close()
			delete(r.workers, addr)
		}
	}
	r.rebuildAddrs()
}

func (r *Registry) rebuildAddrs() {
	r.addrs = make([]string, 0, len(r.workers))
	for addr := range r.workers {
		r.addrs = append(r.addrs, addr)
	}
	if r.nextIdx >= len(r.addrs) {
		r.nextIdx = 0
	}
}

// NextWorker returns the next worker using round-robin.
func (r *Registry) NextWorker() pb.FractalWorkerClient {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := len(r.addrs)
	if n == 0 {
		return nil
	}

	addr := r.addrs[r.nextIdx%n]
	r.nextIdx = (r.nextIdx + 1) % n
	return r.workers[addr].client
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
		w.conn.Close()
	}
}
