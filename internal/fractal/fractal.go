package fractal

// Calculator computes fractal iteration values for a region of the complex plane.
type Calculator interface {
	// Compute returns iteration counts for each pixel in the given region.
	// The result slice has length pixelWidth * pixelHeight, row-major order.
	Compute(params Params) []uint32
}

type Params struct {
	RealMin       float64
	RealMax       float64
	ImagMin       float64
	ImagMax       float64
	PixelWidth    int
	PixelHeight   int
	MaxIterations int
}

// Registry maps fractal type names to their calculators.
var Registry = map[string]Calculator{
	"mandelbrot": &Mandelbrot{},
}
