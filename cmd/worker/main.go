package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "fractal-cluster/internal/gen/fractal"
	"fractal-cluster/internal/worker"

	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 50051, "gRPC listen port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFractalWorkerServer(s, &worker.Server{})

	log.Printf("Worker listening on :%d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
