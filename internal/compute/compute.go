// Package compute defines the transport-agnostic interface that the
// coordinator uses to ask "a worker" to compute a fractal block.
//
// This is the seam that lets the same coordinator code run in two modes:
//   - distributed: Engine is a GRPCEngine that forwards to a remote worker
//   - all-in-one:  Engine is a LocalEngine that calls fractal.Calculator directly
//
// Nothing in this package imports the generated proto types, so the
// dispatcher and registry stay free of gRPC artifacts.
package compute

import "context"

type Request struct {
	FractalType   string
	RealMin       float64
	RealMax       float64
	ImagMin       float64
	ImagMax       float64
	PixelWidth    int
	PixelHeight   int
	MaxIterations int
	BlockID       string
}

type Response struct {
	BlockID     string
	Iterations  []uint32
	PixelWidth  int
	PixelHeight int
}

type Engine interface {
	Compute(ctx context.Context, req Request) (*Response, error)
}
