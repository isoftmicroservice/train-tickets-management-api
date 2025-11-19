package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ticket "github.com/cloudbees/train-ticket-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := ticket.NewTicketServiceClient(conn)
	ctx := context.Background()

	command := os.Args[1]

	switch command {
	case "purchase":
		purchaseTicket(ctx, client, os.Args[2:])
	case "receipt":
		viewReceipt(ctx, client, os.Args[2:])
	case "allocations":
		viewAllocations(ctx, client, os.Args[2:])
	case "remove":
		removeUser(ctx, client, os.Args[2:])
	case "modify":
		modifySeat(ctx, client, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  purchase <first_name> <last_name> <email>")
	fmt.Println("  receipt <jwt_token>")
	fmt.Println("  allocations <jwt_token> [section]")
	fmt.Println("  remove <jwt_token> [email]")
	fmt.Println("  modify <jwt_token> <section> <seat_number> [email]")
}

func purchaseTicket(ctx context.Context, client ticket.TicketServiceClient, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: purchase <first_name> <last_name> <email>")
		return
	}

	req := &ticket.PurchaseTicketRequest{
		FirstName: args[0],
		LastName:  args[1],
		Email:     args[2],
	}

	resp, err := client.PurchaseTicket(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	printReceipt(resp.Receipt)
}

func viewReceipt(ctx context.Context, client ticket.TicketServiceClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: receipt <jwt_token>")
		return
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + args[0],
	}))

	req := &ticket.ViewUserReceiptRequest{}
	resp, err := client.ViewUserReceipt(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	printReceipt(resp.Receipt)
}

func viewAllocations(ctx context.Context, client ticket.TicketServiceClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: allocations <jwt_token> [section]")
		return
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + args[0],
	}))

	req := &ticket.ViewAllocationsRequest{}
	if len(args) > 1 {
		req.Section = args[1]
	}

	resp, err := client.ViewAllocations(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Total allocations: %d\n\n", len(resp.Allocations))
	for _, alloc := range resp.Allocations {
		fmt.Printf("Section: %s, Seat: %d\n", alloc.Section, alloc.SeatNumber)
		fmt.Printf("  User: %s %s (%s)\n\n", alloc.User.FirstName, alloc.User.LastName, alloc.User.Email)
	}
}

func removeUser(ctx context.Context, client ticket.TicketServiceClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: remove <jwt_token> [email]")
		return
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + args[0],
	}))

	req := &ticket.RemoveUserFromTrainRequest{}
	if len(args) > 1 {
		req.Email = args[1]
	}

	resp, err := client.RemoveUserFromTrain(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Success: %s\n", resp.Message)
}

func modifySeat(ctx context.Context, client ticket.TicketServiceClient, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: modify <jwt_token> <section> <seat_number> [email]")
		return
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + args[0],
	}))

	var seatNumber int32
	fmt.Sscanf(args[2], "%d", &seatNumber)

	req := &ticket.ModifyUserSeatRequest{
		Section:    args[1],
		SeatNumber: seatNumber,
	}
	if len(args) > 3 {
		req.Email = args[3]
	}

	resp, err := client.ModifyUserSeat(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Seat modified successfully:")
	printReceipt(resp.Receipt)
}

func printReceipt(receipt *ticket.Receipt) {
	fmt.Println("=== Receipt ===")
	fmt.Printf("From: %s\n", receipt.From)
	fmt.Printf("To: %s\n", receipt.To)
	fmt.Printf("User: %s %s (%s)\n", receipt.User.FirstName, receipt.User.LastName, receipt.User.Email)
	fmt.Printf("Price: $%.2f\n", float64(receipt.PricePaid)/100)
	fmt.Printf("Seat: %s-%d\n", receipt.Seat.Section, receipt.Seat.SeatNumber)
	fmt.Println("==============")
}

