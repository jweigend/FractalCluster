package worker

import (
	"context"
	"log"

	"fractal-cluster/internal/compute"
	pb "fractal-cluster/internal/gen/fractal"
)

// Server is the gRPC adapter that exposes a LocalEngine to remote
// coordinators. All actual computation happens in compute.LocalEngine,
// which is the same code path the all-in-one binary uses.
type Server struct {
	pb.UnimplementedFractalWorkerServer
	engine *compute.LocalEngine
}

func NewServer() *Server {
	return &Server{engine: compute.NewLocalEngine()}
}

func (s *Server) Compute(ctx context.Context, req *pb.ComputeRequest) (*pb.ComputeResponse, error) {
	log.Printf("Computing block %s: %dx%d pixels, max_iter=%d",
		req.BlockId, req.PixelWidth, req.PixelHeight, req.MaxIterations)

	resp, err := s.engine.Compute(ctx, compute.Request{
		FractalType:   req.FractalType,
		RealMin:       req.RealMin,
		RealMax:       req.RealMax,
		ImagMin:       req.ImagMin,
		ImagMax:       req.ImagMax,
		PixelWidth:    int(req.PixelWidth),
		PixelHeight:   int(req.PixelHeight),
		MaxIterations: int(req.MaxIterations),
		BlockID:       req.BlockId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.ComputeResponse{
		BlockId:     resp.BlockID,
		Iterations:  resp.Iterations,
		PixelWidth:  int32(resp.PixelWidth),
		PixelHeight: int32(resp.PixelHeight),
	}, nil
}
