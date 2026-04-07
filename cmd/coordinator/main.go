package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"fractal-cluster/internal/coordinator"
	pb "fractal-cluster/internal/gen/fractal"

	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "8080", "HTTP listen port")
	grpcPort := flag.String("grpc-port", "9090", "gRPC listen port for worker registration")
	webDir := flag.String("web", "web/dist", "Path to frontend build directory")
	flag.Parse()

	registry := coordinator.NewRegistry(30 * time.Second)
	defer registry.Close()

	registry.StartReaper(10 * time.Second)

	// Start gRPC server for worker registration
	go func() {
		lis, err := net.Listen("tcp", ":"+*grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", *grpcPort, err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterFractalCoordinatorServer(grpcServer, coordinator.NewCoordinatorGRPC(registry))
		log.Printf("Coordinator gRPC listening on :%s (worker registration)", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	server := coordinator.NewServer(registry)

	http.HandleFunc("/ws", server.HandleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(*webDir)))

	log.Printf("Coordinator HTTP listening on :%s (serving frontend from %s)", *port, *webDir)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
