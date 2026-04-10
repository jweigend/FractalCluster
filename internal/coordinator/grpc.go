package coordinator

import (
	"context"

	"fractal-cluster/internal/compute"
	pb "fractal-cluster/internal/gen/fractal"
)

// CoordinatorGRPC handles worker self-registration in distributed mode.
// It is the only place that turns a registered address into a GRPCEngine —
// the registry itself stays transport-agnostic.
type CoordinatorGRPC struct {
	pb.UnimplementedFractalCoordinatorServer
	registry *Registry
}

func NewCoordinatorGRPC(registry *Registry) *CoordinatorGRPC {
	return &CoordinatorGRPC{registry: registry}
}

func (c *CoordinatorGRPC) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	engine, err := compute.DialGRPCEngine(req.WorkerAddress)
	if err != nil {
		return &pb.RegisterResponse{Accepted: false}, err
	}
	c.registry.RegisterEngine(req.WorkerAddress, engine, func() { _ = engine.Close() }, false)
	return &pb.RegisterResponse{Accepted: true}, nil
}

func (c *CoordinatorGRPC) Heartbeat(ctx context.Context, req *pb.WorkerHeartbeatRequest) (*pb.WorkerHeartbeatResponse, error) {
	c.registry.RecordHeartbeat(req.WorkerAddress)
	return &pb.WorkerHeartbeatResponse{Acknowledged: true}, nil
}
