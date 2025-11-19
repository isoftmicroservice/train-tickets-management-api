package main

import (
	"log"
	"net"

	ticket "github.com/cloudbees/train-ticket-service/api"
	"github.com/cloudbees/train-ticket-service/internal/service"
	"github.com/cloudbees/train-ticket-service/internal/store"
	"google.golang.org/grpc"
)

func main() {
	// Create store
	s := store.NewStore()

	// Create service
	ticketService := service.NewTicketService(s)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register service
	ticket.RegisterTicketServiceServer(grpcServer, ticketService)

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("gRPC server starting on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

