package coordinator

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	pb "fractal-cluster/internal/gen/fractal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

type WorkerEntry struct {
	Address string `yaml:"address"`
	conn    *grpc.ClientConn
	client  pb.FractalWorkerClient
	healthy bool
}

type Registry struct {
	mu      sync.RWMutex
	workers []*WorkerEntry
	nextIdx int
	stopCh  chan struct{}
}

type workersConfig struct {
	Workers []struct {
		Address string `yaml:"address"`
	} `yaml:"workers"`
}

func NewRegistry(configPath string) (*Registry, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg workersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	r := &Registry{
		stopCh: make(chan struct{}),
	}

	for _, w := range cfg.Workers {
		entry := &WorkerEntry{Address: w.Address}
		r.workers = append(r.workers, entry)
	}

	return r, nil
}

func (r *Registry) Connect() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, w := range r.workers {
		conn, err := grpc.NewClient(w.Address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Printf("Failed to connect to worker %s: %v", w.Address, err)
			continue
		}
		w.conn = conn
		w.client = pb.NewFractalWorkerClient(conn)

		// Verify the worker is actually reachable before marking healthy
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = w.client.Heartbeat(ctx, &pb.HeartbeatRequest{})
		cancel()

		w.healthy = err == nil
		if w.healthy {
			log.Printf("Connected to worker %s", w.Address)
		} else {
			log.Printf("Worker %s not reachable: %v", w.Address, err)
		}
	}
}

func (r *Registry) StartHealthChecks(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-r.stopCh:
				return
			case <-ticker.C:
				r.checkAll()
			}
		}
	}()
}

func (r *Registry) checkAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, w := range r.workers {
		if w.client == nil {
			// Try to reconnect
			conn, err := grpc.NewClient(w.Address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				continue
			}
			w.conn = conn
			w.client = pb.NewFractalWorkerClient(conn)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := w.client.Heartbeat(ctx, &pb.HeartbeatRequest{})
		cancel()

		wasHealthy := w.healthy
		w.healthy = err == nil
		if wasHealthy && !w.healthy {
			log.Printf("Worker %s is now unhealthy", w.Address)
		} else if !wasHealthy && w.healthy {
			log.Printf("Worker %s is now healthy", w.Address)
		}
	}
}

// NextWorker returns the next healthy worker using round-robin.
func (r *Registry) NextWorker() pb.FractalWorkerClient {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := len(r.workers)
	for i := 0; i < n; i++ {
		idx := (r.nextIdx + i) % n
		if r.workers[idx].healthy && r.workers[idx].client != nil {
			r.nextIdx = idx + 1
			return r.workers[idx].client
		}
	}
	return nil
}

func (r *Registry) HealthyCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, w := range r.workers {
		if w.healthy {
			count++
		}
	}
	return count
}

func (r *Registry) Close() {
	close(r.stopCh)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range r.workers {
		if w.conn != nil {
			w.conn.Close()
		}
	}
}
