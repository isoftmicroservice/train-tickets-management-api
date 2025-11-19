package service

import (
	"context"
	"testing"

	ticket "github.com/cloudbees/train-ticket-service/api"
	"github.com/cloudbees/train-ticket-service/internal/config"
	"github.com/cloudbees/train-ticket-service/internal/model"
	"github.com/cloudbees/train-ticket-service/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// createTestJWT creates a valid JWT token for testing
func createTestJWT(email, firstName, lastName, role string) string {
	claims := jwt.MapClaims{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"role":       role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func TestPurchaseTicket(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	req := &ticket.PurchaseTicketRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	resp, err := service.PurchaseTicket(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Receipt == nil {
		t.Fatal("Expected receipt, got nil")
	}

	if resp.Receipt.User.Email != "john@example.com" {
		t.Errorf("Expected email john@example.com, got %s", resp.Receipt.User.Email)
	}

	if resp.Receipt.From != config.RouteFrom {
		t.Errorf("Expected from %s, got %s", config.RouteFrom, resp.Receipt.From)
	}

	if resp.Receipt.PricePaid != config.TicketPriceCents {
		t.Errorf("Expected price %d, got %d", config.TicketPriceCents, resp.Receipt.PricePaid)
	}
}

func TestPurchaseTicket_InvalidInput(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	tests := []struct {
		name string
		req  *ticket.PurchaseTicketRequest
	}{
		{"missing first name", &ticket.PurchaseTicketRequest{LastName: "Doe", Email: "test@example.com"}},
		{"missing last name", &ticket.PurchaseTicketRequest{FirstName: "John", Email: "test@example.com"}},
		{"missing email", &ticket.PurchaseTicketRequest{FirstName: "John", LastName: "Doe"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.PurchaseTicket(context.Background(), tt.req)
			if err == nil {
				t.Error("Expected error for invalid input")
				return
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Error("Expected gRPC status error")
				return
			}

			if st.Code() != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument, got %v", st.Code())
			}
		})
	}
}

func TestPurchaseTicket_Duplicate(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	req := &ticket.PurchaseTicketRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	// First purchase
	_, err := service.PurchaseTicket(context.Background(), req)
	if err != nil {
		t.Fatalf("First purchase failed: %v", err)
	}

	// Duplicate purchase
	_, err = service.PurchaseTicket(context.Background(), req)
	if err == nil {
		t.Error("Expected error for duplicate purchase")
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
		return
	}

	if st.Code() != codes.AlreadyExists {
		t.Errorf("Expected AlreadyExists, got %v", st.Code())
	}
}

func TestViewUserReceipt(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	// Purchase ticket first
	user := model.User{FirstName: "Jane", LastName: "Smith", Email: "jane@example.com"}
	_, err := s.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	// Create context with JWT metadata
	token := createTestJWT("jane@example.com", "Jane", "Smith", "user")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &ticket.ViewUserReceiptRequest{}
	resp, err := service.ViewUserReceipt(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Receipt == nil {
		t.Fatal("Expected receipt, got nil")
	}

	if resp.Receipt.User.Email != "jane@example.com" {
		t.Errorf("Expected email jane@example.com, got %s", resp.Receipt.User.Email)
	}
}

func TestViewUserReceipt_NoAuth(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	ctx := context.Background()
	req := &ticket.ViewUserReceiptRequest{}

	_, err := service.ViewUserReceipt(ctx, req)
	if err == nil {
		t.Error("Expected error for missing auth")
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
		return
	}

	if st.Code() != codes.Unauthenticated {
		t.Errorf("Expected Unauthenticated, got %v", st.Code())
	}
}

func TestViewAllocations_Admin(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	// Purchase some tickets
	user1 := model.User{Email: "user1@example.com", FirstName: "User1", LastName: "One"}
	user2 := model.User{Email: "user2@example.com", FirstName: "User2", LastName: "Two"}
	s.PurchaseTicket(user1, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	s.PurchaseTicket(user2, config.RouteFrom, config.RouteTo, config.TicketPriceCents)

	// Create context with admin JWT
	token := createTestJWT("admin@example.com", "Admin", "User", "admin")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &ticket.ViewAllocationsRequest{}
	resp, err := service.ViewAllocations(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(resp.Allocations) < 2 {
		t.Errorf("Expected at least 2 allocations, got %d", len(resp.Allocations))
	}
}

func TestViewAllocations_NonAdmin(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	// Create context with non-admin JWT
	token := createTestJWT("user@example.com", "User", "Test", "user")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &ticket.ViewAllocationsRequest{}
	_, err := service.ViewAllocations(ctx, req)
	if err == nil {
		t.Error("Expected error for non-admin")
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
		return
	}

	if st.Code() != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", st.Code())
	}
}

func TestRemoveUserFromTrain(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	// Purchase ticket first
	user := model.User{Email: "remove@example.com", FirstName: "Remove", LastName: "Me"}
	_, err := s.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	// Create context with JWT
	token := createTestJWT("remove@example.com", "Remove", "Me", "user")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &ticket.RemoveUserFromTrainRequest{}
	resp, err := service.RemoveUserFromTrain(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	// Verify ticket is removed
	_, err = s.GetTicketByEmail(user.Email)
	if err != store.ErrTicketNotFound {
		t.Error("Expected ticket to be removed")
	}
}

func TestModifyUserSeat(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	// Purchase ticket first
	user := model.User{Email: "modify@example.com", FirstName: "Modify", LastName: "Seat"}
	_, err := s.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	// Create context with JWT
	token := createTestJWT("modify@example.com", "Modify", "Seat", "user")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	req := &ticket.ModifyUserSeatRequest{
		Section:    "B",
		SeatNumber: 5,
	}

	resp, err := service.ModifyUserSeat(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Receipt.Seat.Section != "B" {
		t.Errorf("Expected section B, got %s", resp.Receipt.Seat.Section)
	}

	if resp.Receipt.Seat.SeatNumber != 5 {
		t.Errorf("Expected seat number 5, got %d", resp.Receipt.Seat.SeatNumber)
	}
}

func TestModifyUserSeat_InvalidInput(t *testing.T) {
	s := store.NewStore()
	service := NewTicketService(s)

	token := createTestJWT("test@example.com", "Test", "User", "user")
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	tests := []struct {
		name string
		req  *ticket.ModifyUserSeatRequest
	}{
		{"missing section", &ticket.ModifyUserSeatRequest{SeatNumber: 5}},
		{"missing seat number", &ticket.ModifyUserSeatRequest{Section: "A"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ModifyUserSeat(ctx, tt.req)
			if err == nil {
				t.Error("Expected error for invalid input")
				return
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Error("Expected gRPC status error")
				return
			}

			if st.Code() != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument, got %v", st.Code())
			}
		})
	}
}

