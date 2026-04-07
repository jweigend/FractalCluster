package worker

import (
	"context"
	"log"
	"time"

	pb "fractal-cluster/internal/gen/fractal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Registrar struct {
	coordinatorAddr string
	advertiseAddr   string
	client          pb.FractalCoordinatorClient
	conn            *grpc.ClientConn
	stopCh          chan struct{}
}

func NewRegistrar(coordinatorAddr, advertiseAddr string) *Registrar {
	return &Registrar{
		coordinatorAddr: coordinatorAddr,
		advertiseAddr:   advertiseAddr,
		stopCh:          make(chan struct{}),
	}
}

// Start connects to the coordinator, registers, and begins sending heartbeats.
// Retries registration until successful.
func (r *Registrar) Start() error {
	conn, err := grpc.NewClient(r.coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	r.conn = conn
	r.client = pb.NewFractalCoordinatorClient(conn)

	// Register with retry
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := r.client.Register(ctx, &pb.RegisterRequest{WorkerAddress: r.advertiseAddr})
		cancel()

		if err == nil && resp.Accepted {
			log.Printf("Registered at coordinator %s as %s", r.coordinatorAddr, r.advertiseAddr)
			break
		}
		log.Printf("Registration failed (%v), retrying in 2s...", err)

		select {
		case <-r.stopCh:
			return nil
		case <-time.After(2 * time.Second):
		}
	}

	// Heartbeat loop
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-r.stopCh:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err := r.client.Heartbeat(ctx, &pb.WorkerHeartbeatRequest{WorkerAddress: r.advertiseAddr})
				cancel()
				if err != nil {
					log.Printf("Heartbeat failed: %v, re-registering...", err)
					ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
					r.client.Register(ctx2, &pb.RegisterRequest{WorkerAddress: r.advertiseAddr})
					cancel2()
				}
			}
		}
	}()

	return nil
}

func (r *Registrar) Stop() {
	close(r.stopCh)
	if r.conn != nil {
		r.conn.Close()
	}
}
