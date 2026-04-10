package compute

import (
	"context"

	"fractal-cluster/internal/fractal"
)

// LocalEngine runs fractal computation in the current process. It is the
// canonical place where fractal.Calculator is invoked: both the all-in-one
// binary and the distributed worker's gRPC handler delegate here.
type LocalEngine struct{}

func NewLocalEngine() *LocalEngine { return &LocalEngine{} }

func (e *LocalEngine) Compute(_ context.Context, req Request) (*Response, error) {
	calc, ok := fractal.Registry[req.FractalType]
	if !ok {
		calc = fractal.Registry["mandelbrot"]
	}
	iters := calc.Compute(fractal.Params{
		RealMin:       req.RealMin,
		RealMax:       req.RealMax,
		ImagMin:       req.ImagMin,
		ImagMax:       req.ImagMax,
		PixelWidth:    req.PixelWidth,
		PixelHeight:   req.PixelHeight,
		MaxIterations: req.MaxIterations,
	})
	return &Response{
		BlockID:     req.BlockID,
		Iterations:  iters,
		PixelWidth:  req.PixelWidth,
		PixelHeight: req.PixelHeight,
	}, nil
}
