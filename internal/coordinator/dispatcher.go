package coordinator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	pb "fractal-cluster/internal/gen/fractal"
	"fractal-cluster/internal/protocol"
)

type Dispatcher struct {
	registry *Registry
}

func NewDispatcher(registry *Registry) *Dispatcher {
	return &Dispatcher{registry: registry}
}

type DispatchResult struct {
	Block  Block
	Result *pb.ComputeResponse
	Err    error
}

// Dispatch distributes blocks to workers and sends results back via the callback.
// It respects maxThreads as the concurrency limit.
// Returns when all blocks are done or ctx is cancelled.
func (d *Dispatcher) Dispatch(ctx context.Context, params protocol.CalcParams, blocks []Block, onStarted func(Block), onResult func(DispatchResult, int, int)) {
	sem := make(chan struct{}, params.MaxThreads)
	var wg sync.WaitGroup
	var completed atomic.Int32
	total := len(blocks)

	for _, block := range blocks {
		select {
		case <-ctx.Done():
			return
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(b Block) {
			defer wg.Done()
			defer func() { <-sem }()

			onStarted(b)
			result := d.computeBlock(ctx, params, b)
			c := int(completed.Add(1))
			onResult(result, c, total)
		}(block)
	}

	wg.Wait()
}

func (d *Dispatcher) computeBlock(ctx context.Context, params protocol.CalcParams, block Block) DispatchResult {
	// Map block pixel coordinates to complex plane coordinates
	scaleR := (params.RealMax - params.RealMin) / float64(params.PicWidth)
	scaleI := (params.ImagMax - params.ImagMin) / float64(params.PicHeight)

	req := &pb.ComputeRequest{
		FractalType:   params.FractalType,
		RealMin:       params.RealMin + float64(block.X)*scaleR,
		RealMax:       params.RealMin + float64(block.X+block.Width)*scaleR,
		ImagMin:       params.ImagMin + float64(block.Y)*scaleI,
		ImagMax:       params.ImagMin + float64(block.Y+block.Height)*scaleI,
		PixelWidth:    int32(block.Width),
		PixelHeight:   int32(block.Height),
		MaxIterations: int32(params.MaxIterations),
		BlockId:       block.ID,
	}

	// Note: ImagMin/ImagMax need to be swapped because screen Y is inverted
	// In the complex plane, higher Y = higher imaginary, but on screen higher Y = lower
	req.ImagMin, req.ImagMax = params.ImagMax-float64(block.Y+block.Height)*scaleI, params.ImagMax-float64(block.Y)*scaleI

	// Try to get a worker, retry on failure
	for attempt := 0; attempt < 3; attempt++ {
		select {
		case <-ctx.Done():
			return DispatchResult{Block: block, Err: ctx.Err()}
		default:
		}

		worker := d.registry.NextWorker()
		if worker == nil {
			log.Printf("No healthy workers available for block %s", block.ID)
			continue
		}

		callCtx, callCancel := context.WithTimeout(ctx, 30*time.Second)
		resp, err := worker.Compute(callCtx, req)
		callCancel()
		if err != nil {
			log.Printf("Worker failed for block %s (attempt %d): %v", block.ID, attempt+1, err)
			continue
		}

		return DispatchResult{Block: block, Result: resp}
	}

	return DispatchResult{Block: block, Err: fmt.Errorf("all attempts failed for block %s", block.ID)}
}
