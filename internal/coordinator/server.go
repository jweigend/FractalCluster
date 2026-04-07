package coordinator

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"fractal-cluster/internal/protocol"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	registry   *Registry
	dispatcher *Dispatcher
	mu         sync.Mutex
	cancelFn   context.CancelFunc
}

func NewServer(registry *Registry) *Server {
	return &Server{
		registry:   registry,
		dispatcher: NewDispatcher(registry),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		var envelope struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "start":
			var startMsg protocol.StartMessage
			if err := json.Unmarshal(msg, &startMsg); err != nil {
				s.sendError(conn, "Invalid start message")
				continue
			}
			s.handleStart(conn, startMsg.Params)

		case "stop":
			s.handleStop()
		}
	}
}

func (s *Server) handleStart(conn *websocket.Conn, params protocol.CalcParams) {
	s.handleStop() // Cancel any running calculation

	if params.FractalType == "" {
		params.FractalType = "mandelbrot"
	}
	if params.MaxThreads <= 0 {
		params.MaxThreads = 10
	}
	if params.MaxIterations <= 0 {
		params.MaxIterations = 100
	}

	var blocks []Block
	if params.DistActive {
		blocks = SplitImage(params.PicWidth, params.PicHeight, params.BlockWidth, params.BlockHeight)
	} else {
		blocks = []Block{{ID: "0-0", X: 0, Y: 0, Width: params.PicWidth, Height: params.PicHeight}}
	}

	log.Printf("Starting calculation: %dx%d image, %d blocks, max_iter=%d, workers=%d",
		params.PicWidth, params.PicHeight, len(blocks), params.MaxIterations, s.registry.HealthyCount())

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancelFn = cancel
	s.mu.Unlock()

	wsMu := &sync.Mutex{}

	go s.dispatcher.Dispatch(ctx, params, blocks, func(block Block) {
		wsMu.Lock()
		defer wsMu.Unlock()

		s.sendJSON(conn, protocol.BlockStartedMessage{
			Type:    "block_started",
			BlockID: block.ID,
			X:       block.X,
			Y:       block.Y,
			Width:   block.Width,
			Height:  block.Height,
		})
	}, func(result DispatchResult, completed, total int) {
		wsMu.Lock()
		defer wsMu.Unlock()

		if result.Err != nil {
			s.sendJSON(conn, protocol.ErrorMessage{
				Type:    "error",
				Message: result.Err.Error(),
			})
			return
		}

		s.sendJSON(conn, protocol.BlockResultMessage{
			Type:       "block_result",
			BlockID:    result.Block.ID,
			X:          result.Block.X,
			Y:          result.Block.Y,
			Width:      result.Block.Width,
			Height:     result.Block.Height,
			Iterations: result.Result.Iterations,
		})

		s.sendJSON(conn, protocol.ProgressMessage{
			Type:      "progress",
			Completed: completed,
			Total:     total,
		})
	})
}

func (s *Server) handleStop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFn != nil {
		s.cancelFn()
		s.cancelFn = nil
	}
}

func (s *Server) sendJSON(conn *websocket.Conn, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}

func (s *Server) sendError(conn *websocket.Conn, msg string) {
	s.sendJSON(conn, protocol.ErrorMessage{Type: "error", Message: msg})
}
