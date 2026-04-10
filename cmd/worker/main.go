package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "fractal-cluster/internal/gen/fractal"
	"fractal-cluster/internal/worker"

	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 50051, "gRPC listen port")
	coordinator := flag.String("coordinator", "localhost:9090", "Coordinator gRPC address")
	advertise := flag.String("advertise", "", "Address to advertise to coordinator (default: hostname:port)")
	flag.Parse()

	addr := *advertise
	if addr == "" {
		hostname, _ := os.Hostname()
		addr = fmt.Sprintf("%s:%d", hostname, *port)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFractalWorkerServer(s, worker.NewServer())

	log.Printf("Worker listening on :%d", *port)

	// Register at coordinator
	registrar := worker.NewRegistrar(*coordinator, addr)
	go func() {
		if err := registrar.Start(); err != nil {
			log.Printf("Registrar failed: %v", err)
		}
	}()
	defer registrar.Stop()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
