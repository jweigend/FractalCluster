package fractal

import (
	"runtime"
	"sync"
)

type Mandelbrot struct{}

func (m *Mandelbrot) Compute(p Params) []uint32 {
	w, h := p.PixelWidth, p.PixelHeight
	result := make([]uint32, w*h)

	scaleR := (p.RealMax - p.RealMin) / float64(w)
	scaleI := (p.ImagMax - p.ImagMin) / float64(h)

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	rowCh := make(chan int, h)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range rowCh {
				ci := p.ImagMax - float64(y)*scaleI
				offset := y * w
				for x := 0; x < w; x++ {
					cr := p.RealMin + float64(x)*scaleR
					result[offset+x] = mandelbrotIter(cr, ci, p.MaxIterations)
				}
			}
		}()
	}

	for y := 0; y < h; y++ {
		rowCh <- y
	}
	close(rowCh)
	wg.Wait()

	return result
}

func mandelbrotIter(cr, ci float64, maxIter int) uint32 {
	var zr, zi float64
	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi
		if zr2+zi2 > 4.0 {
			return uint32(i)
		}
		zi = 2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return uint32(maxIter)
}
