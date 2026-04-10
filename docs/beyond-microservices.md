# Beyond Microservices: Why a 90s Discipline Deserves a Comeback

*How a 700-line fractal renderer reminded me of a pattern the microservice generation forgot — and why it's no longer controversial to say so.*

---

## A confession from ten years ago

A decade ago I used to argue with people about microservices. Not about whether they were *useful* — that was obvious for some problems — but about whether you should always start with them. I kept saying the same thing: build your application so it can also run as a single process. Keep the business logic clean. Put a thin transport seam between your components. Then, when you actually need to distribute, you flip a factory and you're distributed. Until then, you debug, test, and develop in one process, with one stack trace, in one debugger.

The reaction was usually religious. "That's not cloud-native." "You're coupling your services." "Real systems run in Kubernetes." I gave up trying to convince anyone.

Ten years later, the conversation has shifted. Nobody seriously disputes that ten distributed processes are painful to debug, that local dev environments for microservice stacks are a nightmare, or that most "services" in a typical microservice codebase exist for reasons that have nothing to do with the business logic — auth proxies, sidecars, API gateways, message brokers, service meshes. The technical scaffolding has eaten the application.

So this is an article I would have written in 2015 if anyone had wanted to read it. Today it's not controversial anymore. It's just good engineering hygiene that fell out of fashion and deserves to come back.

I'll make the case with a small, complete example: a distributed fractal renderer written in Go, with a coordinator, workers, a web frontend, and gRPC between them. About 700 lines of Go. Then I'll show how making it *also* run as a single process took roughly an hour, didn't compromise the distributed version at all, and left the codebase clearer than before.

---

## The example: a distributed fractal renderer

The application is deliberately small but architecturally honest. It's the kind of thing that, in a microservice tutorial, would already be split into three repos and a Helm chart.

**The components:**

- A **coordinator** that accepts a render request from the browser, splits the image into blocks, dispatches the blocks to workers, collects results, and streams them back to the frontend over WebSocket.
- One or more **workers** that compute the iteration counts for a block of the complex plane (Mandelbrot, in this case) and return them.
- A **web frontend** (React + Canvas) that lets you pan, zoom, and watch blocks fill in as workers complete them.
- **gRPC** for coordinator–worker communication, defined in a `.proto` file.
- A **dynamic registration** protocol: workers self-register at startup and send heartbeats; the coordinator reaps them on timeout.

The whole thing is the kind of application a textbook would call "a small distributed system." In Go it weighs about 730 lines, plus some React on the frontend.

This is also, charmingly, a rewrite of a project I built in 1998 at FH Rosenheim using Visual Basic 6 and COM+/DCOM. Same architecture: a dispatcher, worker nodes, network IPC, a GUI client. Different transport, different language, same shape. The fact that the shape survives twenty-eight years of technology change is itself a hint that the *architecture* is durable and the *transport* is incidental — which is exactly the lesson this article is about.

## How you'd build it today

If you started this project on a fresh laptop in 2026, here's the path you'd be nudged toward by every tutorial, blog post, and starter template:

1. Define a `proto/fractal.proto` with the gRPC services.
2. Generate Go stubs.
3. Write `cmd/coordinator/main.go` that listens on an HTTP port for the frontend and a gRPC port for worker registration.
4. Write `cmd/worker/main.go` that listens on its own gRPC port and registers with the coordinator at startup.
5. Write a `docker/Dockerfile.coordinator` and a `docker/Dockerfile.worker`.
6. Write a `docker-compose.yaml` that brings up one coordinator and two workers, with a network between them.
7. `make docker` and watch it come up.

This is exactly what fractal-cluster looked like when I rewrote it. And honestly, it's a perfectly reasonable starting point. Docker Compose for two workers and a coordinator is fine. You can `docker compose logs -f` and see what's happening. You can scale the workers with `--scale`. When it works, it feels professional.

But notice what's already true even at this scale:

- You have **three OS processes** (or more) talking over the network, even when developing on one laptop.
- A breakpoint in the worker doesn't pause the coordinator. A breakpoint in the coordinator doesn't pause the worker.
- A stack trace from a failing render starts in the WebSocket handler, hits a wall at the gRPC client boundary, and resumes — *if you're lucky* — in another process's logs that you have to correlate by request ID.
- To run the app at all, you need Docker, or you need three terminals, or you need a `tmuxinator` config, or you need a `Procfile`.
- Editing the business logic means rebuilding a container or restarting two processes.

At three processes, all of this is *annoying but manageable*. At ten processes — which is where most "real" microservice applications land — it's chaos. And here's the part that bothers me most: the chaos has almost nothing to do with the actual business problem you're trying to solve.

## The hidden cost: technical surrogates eat the application

Look at any mature microservice stack and count what's actually doing application work versus what's plumbing. You'll usually find something like:

- One or two services that contain the **actual domain logic** — the thing the business cares about.
- An API gateway.
- An auth service.
- A service mesh sidecar in every pod.
- A message broker for async events.
- A schema registry.
- A config service.
- A secrets manager.
- A distributed tracing collector, because otherwise you can't debug anything.
- A handful of "BFF" (backend-for-frontend) services that exist because the API gateway couldn't quite do what the frontend needed.

Most of these exist *because the system is distributed*, not because the business needs them. They are technical surrogates. They solve problems that wouldn't exist if the same logic ran in one process. Auth between services is hard because the services are separate. Distributed tracing exists because stack traces stop at process boundaries. The schema registry exists because two services serialize the same data and might disagree about the format.

None of this is wrong when you genuinely need the distribution. Some of these problems are unavoidable at real scale, with real teams, real deployment topologies, real failure domains. But here's the question almost nobody asks early enough:

**Does my business logic care that this is distributed?**

For the fractal renderer, the answer is: not at all. The Mandelbrot calculation doesn't know whether the function that called it lives in the same process or arrived over gRPC. The coordinator's job — split image, dispatch blocks, collect results, stream to frontend — is the same regardless of where the workers live. The *only* thing distribution gives you here is the ability to run more CPUs against one render. That's a deployment concern, not a design concern.

And yet, in the distributed version, the gRPC client type leaks into the dispatcher. The `pb.ComputeRequest` proto type appears in the same file as the block-splitting logic. The registry holds `pb.FractalWorkerClient` instances directly. None of that has to do with fractals. All of it makes the code harder to read and harder to debug.

## What the 90s knew

In the late 90s, when COM+ and CORBA were the way you built distributed applications, there was a discipline that the microservice generation largely forgot: you designed your components so they could run *in-process or out-of-process* without changing the calling code. COM had this baked in — an in-proc server and an out-of-proc server looked identical to the client. You could literally flip a registry setting.

The reason wasn't ideological. It was practical: distributed debugging was so painful that nobody would do it if they could avoid it. So you developed against the in-proc version, where the debugger worked, where breakpoints stopped the world, where stack traces went all the way down. You only switched to the distributed version when you needed to deploy across machines. The transport was a deployment concern, not a development concern.

Then containers happened, microservices became fashionable, and we lost this discipline entirely. The default became: distributed from day one, debugged by log correlation, developed with eight terminals open. We accepted enormous tax on our daily workflow in exchange for an architectural property we usually didn't need yet.

I think the discipline is worth reactivating. And I think it's much easier than people assume.

## The seam: one interface, two implementations, three binaries

Here's how the fractal-cluster looks after the refactor. The whole thing rests on a single interface:

```go
// internal/compute/compute.go
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
```

Notice what's *not* in this file: no `import "google.golang.org/grpc"`, no proto types, no network code. Just plain Go types describing "give me a region of the complex plane, get back iteration counts." This is the seam.

There are two implementations.

**`LocalEngine`** runs in the same process. It's the canonical place where the fractal calculator gets called:

```go
func (e *LocalEngine) Compute(_ context.Context, req Request) (*Response, error) {
    calc := fractal.Registry[req.FractalType]
    iters := calc.Compute(fractal.Params{ /* ... */ })
    return &Response{BlockID: req.BlockID, Iterations: iters, /* ... */}, nil
}
```

**`GRPCEngine`** wraps a gRPC client. This is the *only* file on the coordinator side that knows proto types exist:

```go
func (e *GRPCEngine) Compute(ctx context.Context, req Request) (*Response, error) {
    resp, err := e.client.Compute(ctx, &pb.ComputeRequest{ /* translate */ })
    if err != nil {
        return nil, err
    }
    return &Response{ /* translate back */ }, nil
}
```

The dispatcher — the thing that splits the image into blocks and sends them to workers — now talks only to the `Engine` interface. It has no idea whether its workers are in-process or remote:

```go
engine := d.registry.NextEngine()
resp, err := engine.Compute(callCtx, req)
```

And there are now three binaries instead of two:

- `cmd/coordinator` — the distributed coordinator, exactly as before. Its registry is fed `GRPCEngine` instances when workers self-register over gRPC.
- `cmd/worker` — the distributed worker. Internally, the gRPC handler is now a thin adapter that delegates to `LocalEngine`. The same code path runs in both modes.
- `cmd/allinone` — the new binary. It creates a `Registry`, registers a single `LocalEngine`, starts the WebSocket server, and that's it. No gRPC server. No worker registrar. No heartbeat loop. No proto types in the running process.

Here's the entire all-in-one main:

```go
func main() {
    port := flag.String("port", "8080", "HTTP listen port")
    webDir := flag.String("web", "web/dist", "Frontend build directory")
    flag.Parse()

    registry := coordinator.NewRegistry(30 * time.Second)
    defer registry.Close()

    registry.RegisterEngine("local", compute.NewLocalEngine(), nil, true)

    server := coordinator.NewServer(registry)
    http.HandleFunc("/ws", server.HandleWebSocket)
    http.Handle("/", http.FileServer(http.Dir(*webDir)))

    log.Fatal(http.ListenAndServe(":"+*port, nil))
}
```

That's it. Run it with `make run-allinone`, open a browser, and you have the entire application — coordinator logic, worker logic, web server, fractal calculation — in one process. One PID. One stack trace from `HandleWebSocket` all the way down to `mandelbrotIter`. One breakpoint stops everything.

## What it cost

I'll be specific, because the perceived cost is the main reason people don't do this.

The refactor took about an hour. It produced:

- Three new files in `internal/compute/` (compute.go, local.go, grpc.go), totaling about 100 lines.
- A modified `Registry` that holds `Engine` instead of `pb.FractalWorkerClient`. Same number of lines, plus a `local` flag that exempts in-process workers from the heartbeat reaper.
- A modified `Dispatcher` that builds `compute.Request` instead of `pb.ComputeRequest`. Same shape, fewer imports.
- A worker server that's now a thin gRPC adapter around `LocalEngine` — *the worker is now smaller*, not bigger.
- One new `cmd/allinone/main.go`.

The distributed binaries didn't lose any functionality. The proto file didn't change. The Docker setup didn't change. The frontend didn't change.

**The codebase is objectively cleaner than it was.** The dispatcher used to import gRPC types. Now it doesn't. The registry used to dial gRPC connections itself. Now the dialing lives in one place, behind one factory function. The seam I introduced for the all-in-one binary also happened to be the right architectural seam for the distributed version. This is not a coincidence — it almost never is.

## What you get

The list is short and the items are large.

**You can debug.** A real debugger, with breakpoints that stop everything, with a stack trace from the WebSocket frame down to the iteration loop. No log correlation. No distributed tracing.

**You can develop without infrastructure.** No Docker. No Compose. No three terminals. `go run ./cmd/allinone` and you're working.

**Your tests get easier.** An integration test for the dispatcher used to need two processes and a gRPC connection. Now it constructs a `Registry`, registers a `LocalEngine`, and asserts on the result — in-process, in-memory, microseconds.

**Your business logic can't accidentally depend on transport.** The compiler enforces it: the dispatcher doesn't import the proto package. If somebody tries to sneak in a `pb.ComputeRequest`, the all-in-one build catches it.

**You retain everything the distributed version gives you.** Multiple workers, network fanout, dynamic registration, heartbeats, container deployment — all still there, identical, unchanged. You haven't given up *anything*. You've added an option.

**Newcomers can read the code.** Three small `main.go` files showing the topology variants is an extraordinarily clear way to teach a distributed system. "Here's the version with no network. Here's the version with one network hop. Here's how the dispatcher doesn't know the difference." A student can step through the all-in-one version, understand the algorithm, and *then* learn how gRPC fits in. That's a much gentler ramp than dropping someone into a Compose stack.

## When distribution actually pays for itself

I'm not arguing against distributed systems. I'm arguing against *premature* distribution — distributing things before you have a reason to, and paying the daily debugging tax for an architectural property you're not yet using.

Distribute when you have a real reason: independent scaling, fault isolation, separate deployment cadences across teams, hard physical constraints (the data lives over there), regulatory boundaries. These are real reasons and they justify real cost.

But don't distribute because the tutorial said to. Don't distribute because microservices are the default. Don't pay for distributed tracing, service mesh, sidecars, schema registries, and an eight-terminal dev environment when your business logic would happily run in one process and let you debug it with a single breakpoint.

The discipline is simple: **make the seam, then choose the topology.** When the business logic is correct and you understand the system, *then* worry about distribution. The 90s knew this. The microservice generation forgot it. We can remember it again. The cost is an hour and one extra `main.go`. The payoff is the rest of your debugging life.

---

*The code in this article is from [fractal-cluster](https://github.com/jweigend/fractal-cluster), a small open-source rewrite of a 1998 COM+/DCOM project. Run `make run-allinone` for the single-process version, or `make docker` for the distributed one. The dispatcher and the fractal calculation are exactly the same in both.*
