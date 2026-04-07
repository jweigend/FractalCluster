package coordinator

import (
	"context"

	pb "fractal-cluster/internal/gen/fractal"
)

type CoordinatorGRPC struct {
	pb.UnimplementedFractalCoordinatorServer
	registry *Registry
}

func NewCoordinatorGRPC(registry *Registry) *CoordinatorGRPC {
	return &CoordinatorGRPC{registry: registry}
}

func (c *CoordinatorGRPC) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := c.registry.Register(req.WorkerAddress)
	if err != nil {
		return &pb.RegisterResponse{Accepted: false}, err
	}
	return &pb.RegisterResponse{Accepted: true}, nil
}

func (c *CoordinatorGRPC) Heartbeat(ctx context.Context, req *pb.WorkerHeartbeatRequest) (*pb.WorkerHeartbeatResponse, error) {
	c.registry.RecordHeartbeat(req.WorkerAddress)
	return &pb.WorkerHeartbeatResponse{Acknowledged: true}, nil
}
