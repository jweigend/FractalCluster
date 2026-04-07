package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"fractal-cluster/internal/coordinator"
)

func main() {
	port := flag.String("port", "8080", "HTTP listen port")
	workersFile := flag.String("workers", "workers.yaml", "Path to workers config file")
	webDir := flag.String("web", "web/dist", "Path to frontend build directory")
	flag.Parse()

	registry, err := coordinator.NewRegistry(*workersFile)
	if err != nil {
		log.Fatalf("Failed to load workers config: %v", err)
	}
	defer registry.Close()

	registry.Connect()
	registry.StartHealthChecks(5 * time.Second)

	server := coordinator.NewServer(registry)

	http.HandleFunc("/ws", server.HandleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(*webDir)))

	log.Printf("Coordinator listening on :%s (serving frontend from %s)", *port, *webDir)
	log.Printf("Healthy workers: %d", registry.HealthyCount())

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
