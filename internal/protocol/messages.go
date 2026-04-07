package protocol

// Client → Coordinator messages
type StartMessage struct {
	Type   string      `json:"type"` // "start"
	Params CalcParams  `json:"params"`
}

type StopMessage struct {
	Type string `json:"type"` // "stop"
}

type CalcParams struct {
	RealMin       float64 `json:"realMin"`
	RealMax       float64 `json:"realMax"`
	ImagMin       float64 `json:"imagMin"`
	ImagMax       float64 `json:"imagMax"`
	PicWidth      int     `json:"picWidth"`
	PicHeight     int     `json:"picHeight"`
	BlockWidth    int     `json:"blockWidth"`
	BlockHeight   int     `json:"blockHeight"`
	MaxIterations int     `json:"maxIterations"`
	MaxThreads    int     `json:"maxThreads"`
	FractalType   string  `json:"fractalType"`
	DistActive    bool    `json:"distActive"`
}

// Coordinator → Client messages
type BlockStartedMessage struct {
	Type    string `json:"type"` // "block_started"
	BlockID string `json:"blockId"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

type BlockResultMessage struct {
	Type       string   `json:"type"` // "block_result"
	BlockID    string   `json:"blockId"`
	X          int      `json:"x"`
	Y          int      `json:"y"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Iterations []uint32 `json:"iterations"`
}

type ProgressMessage struct {
	Type      string `json:"type"` // "progress"
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

type ErrorMessage struct {
	Type    string `json:"type"` // "error"
	Message string `json:"message"`
}
