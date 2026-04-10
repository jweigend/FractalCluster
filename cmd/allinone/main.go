// Command allinone runs the entire fractal-cluster stack — coordinator,
// worker, and webserver — inside a single OS process. No gRPC, no
// network registration, no heartbeats.
//
// This exists for didactic clarity and easier debugging: a single binary
// you can step through end-to-end. The distributed binaries (cmd/coordinator
// and cmd/worker) reuse the exact same dispatcher and fractal code via the
// compute.Engine interface — the only thing that changes between modes is
// which Engine implementation the registry hands out.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"fractal-cluster/internal/compute"
	"fractal-cluster/internal/coordinator"
)

func main() {
	port := flag.String("port", "8080", "HTTP listen port")
	webDir := flag.String("web", "web/dist", "Path to frontend build directory")
	flag.Parse()

	registry := coordinator.NewRegistry(30 * time.Second)
	defer registry.Close()

	// One in-process worker. Marked local=true so the reaper never touches it.
	registry.RegisterEngine("local", compute.NewLocalEngine(), nil, true)

	server := coordinator.NewServer(registry)
	http.HandleFunc("/ws", server.HandleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(*webDir)))

	log.Printf("All-in-one server listening on :%s (frontend from %s)", *port, *webDir)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
