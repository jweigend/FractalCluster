package compute

import (
	"context"

	pb "fractal-cluster/internal/gen/fractal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCEngine forwards Compute calls to a remote worker over gRPC. It is
// the only place on the coordinator side that touches the generated proto
// types — everything upstream (registry, dispatcher) sees only Engine.
type GRPCEngine struct {
	conn   *grpc.ClientConn
	client pb.FractalWorkerClient
}

func DialGRPCEngine(address string) (*GRPCEngine, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &GRPCEngine{conn: conn, client: pb.NewFractalWorkerClient(conn)}, nil
}

func (e *GRPCEngine) Compute(ctx context.Context, req Request) (*Response, error) {
	resp, err := e.client.Compute(ctx, &pb.ComputeRequest{
		FractalType:   req.FractalType,
		RealMin:       req.RealMin,
		RealMax:       req.RealMax,
		ImagMin:       req.ImagMin,
		ImagMax:       req.ImagMax,
		PixelWidth:    int32(req.PixelWidth),
		PixelHeight:   int32(req.PixelHeight),
		MaxIterations: int32(req.MaxIterations),
		BlockId:       req.BlockID,
	})
	if err != nil {
		return nil, err
	}
	return &Response{
		BlockID:     resp.BlockId,
		Iterations:  resp.Iterations,
		PixelWidth:  int(resp.PixelWidth),
		PixelHeight: int(resp.PixelHeight),
	}, nil
}

func (e *GRPCEngine) Close() error { return e.conn.Close() }
