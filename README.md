# Fractal Cluster

A distributed fractal renderer built with Go, gRPC, and React. Originally inspired by a VB6/COM+ demo from 1996 that showcased distributed computing across networked Windows machines, this project reimagines the same concept with a modern stack.

![Screenshot](images/Screenshot.png)

## Architecture

The system follows a coordinator/worker pattern. The browser connects via WebSocket to a coordinator, which splits the image into blocks and distributes them to workers over gRPC. Results stream back block-by-block, rendering progressively in the browser.

```
Browser (React + Canvas)
   |
   | WebSocket (JSON)
   |
Coordinator (Go)
   |
   | gRPC (Protobuf)
   |
 Workers (Go, n instances)
```

### Components

**Frontend** --- React, TypeScript, Vite, Canvas 2D

The UI sends calculation parameters (complex plane bounds, resolution, iteration depth) over a WebSocket connection. As the coordinator dispatches blocks, the frontend receives `block_started` events (red outlines showing parallelism) followed by `block_result` events (pixel data rendered via `putImageData`). Zoom is handled by mouse drag: left-click zooms in, right-click zooms out.

**Coordinator** --- Go, gorilla/websocket

Receives calculation requests from the browser and orchestrates the computation:

1. **Splitter** divides the image into rectangular blocks (e.g. 30x30 px), merging remainder blocks that would be too small.
2. **Registry** maintains persistent gRPC connections to all workers, runs periodic health checks (heartbeat RPC), and provides round-robin load balancing.
3. **Dispatcher** distributes blocks to healthy workers with a semaphore controlling max concurrency. Failed blocks are retried up to 3 times on different workers.

**Worker** --- Go, gRPC

A stateless compute node. Receives a `ComputeRequest` (complex plane region, pixel dimensions, max iterations), computes the fractal, and returns an array of iteration counts. Workers parallelize internally across CPU cores. The fractal algorithm is pluggable via a registry; currently Mandelbrot is implemented.

### gRPC Protocol

Defined in [`proto/fractal.proto`](proto/fractal.proto):

| Service | RPC | Purpose |
|---------|-----|---------|
| FractalWorker | `Compute` | Compute iteration values for a rectangular block |
| FractalWorker | `Heartbeat` | Health check |

### WebSocket Protocol

| Direction | Message | Purpose |
|-----------|---------|---------|
| Client -> Server | `start` | Begin calculation with params |
| Client -> Server | `stop` | Cancel running calculation |
| Server -> Client | `block_started` | Block dispatched to worker (triggers outline) |
| Server -> Client | `block_result` | Pixel data for completed block |
| Server -> Client | `progress` | Completed/total block count |

## Running

### Local

Start one or more workers, then the coordinator:

```bash
make build
bin/worker -port 50051 &
bin/worker -port 50052 &
bin/coordinator -port 8080 -workers workers.yaml -web web/dist
```

Worker addresses are configured in [`workers.yaml`](workers.yaml).

Open http://localhost:8080 in a browser.

### Docker

```bash
docker compose up --build
```

This starts a coordinator and two workers. The coordinator is exposed on port 8080. Worker addresses are resolved via Docker DNS ([`workers-docker.yaml`](workers-docker.yaml)).

## Project Structure

```
cmd/
  coordinator/       Coordinator entry point
  worker/            Worker entry point
internal/
  coordinator/       WebSocket server, dispatcher, registry, splitter
  worker/            gRPC server implementation
  fractal/           Fractal algorithms (Mandelbrot) and calculator registry
  gen/fractal/       Generated protobuf/gRPC code
  protocol/          Shared WebSocket message types
proto/
  fractal.proto      gRPC service definition
web/
  src/
    components/      React components (Canvas, ParameterPanel, StatusBar)
    lib/             WebSocket client, canvas renderer, color table
```

## 1996 vs 2026

| | VB6/COM+ (1996) | Go/gRPC/React (2026) |
|---|---|---|
| Compute nodes | COM+ objects on Windows machines | gRPC workers (any OS, containerized) |
| Communication | DCOM (binary, Windows-only) | gRPC + Protobuf (cross-platform, HTTP/2) |
| Work distribution | Manual partitioning | Coordinator with round-robin, health checks, retry |
| Frontend | VB6 Forms, GDI | React + Canvas 2D, WebSocket streaming |
| Deployment | Install on each machine | `docker compose up` |
| Scaling | Add more Windows boxes | Add worker containers |
