package worker

import (
	"context"
	"log"

	pb "fractal-cluster/internal/gen/fractal"
	"fractal-cluster/internal/fractal"
)

type Server struct {
	pb.UnimplementedFractalWorkerServer
}

func (s *Server) Compute(ctx context.Context, req *pb.ComputeRequest) (*pb.ComputeResponse, error) {
	calc, ok := fractal.Registry[req.FractalType]
	if !ok {
		calc = fractal.Registry["mandelbrot"]
	}

	log.Printf("Computing block %s: %dx%d pixels, max_iter=%d",
		req.BlockId, req.PixelWidth, req.PixelHeight, req.MaxIterations)

	params := fractal.Params{
		RealMin:       req.RealMin,
		RealMax:       req.RealMax,
		ImagMin:       req.ImagMin,
		ImagMax:       req.ImagMax,
		PixelWidth:    int(req.PixelWidth),
		PixelHeight:   int(req.PixelHeight),
		MaxIterations: int(req.MaxIterations),
	}

	iterations := calc.Compute(params)

	return &pb.ComputeResponse{
		BlockId:    req.BlockId,
		Iterations: iterations,
		PixelWidth: req.PixelWidth,
		PixelHeight: req.PixelHeight,
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return &pb.HeartbeatResponse{Healthy: true}, nil
}
