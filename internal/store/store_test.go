package store

import (
	"fmt"
	"testing"

	"github.com/cloudbees/train-ticket-service/internal/config"
	"github.com/cloudbees/train-ticket-service/internal/model"
)

func TestPurchaseTicket(t *testing.T) {
	store := NewStore()

	user := model.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	// Test successful purchase
	ticket, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ticket == nil {
		t.Fatal("Expected ticket, got nil")
	}

	if ticket.User.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, ticket.User.Email)
	}

	if ticket.From != config.RouteFrom {
		t.Errorf("Expected from %s, got %s", config.RouteFrom, ticket.From)
	}

	if ticket.PricePaid != config.TicketPriceCents {
		t.Errorf("Expected price %d, got %d", config.TicketPriceCents, ticket.PricePaid)
	}

	// Verify seat is assigned
	if ticket.Seat.Section != "A" && ticket.Seat.Section != "B" {
		t.Errorf("Expected section A or B, got %s", ticket.Seat.Section)
	}

	if ticket.Seat.SeatNumber < 1 || ticket.Seat.SeatNumber > 10 {
		t.Errorf("Expected seat number 1-10, got %d", ticket.Seat.SeatNumber)
	}

	// Test duplicate purchase
	_, err = store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != ErrUserAlreadyHasTicket {
		t.Errorf("Expected ErrUserAlreadyHasTicket, got: %v", err)
	}
}

func TestGetTicketByEmail(t *testing.T) {
	store := NewStore()

	user := model.User{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@example.com",
	}

	// Purchase ticket first
	_, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	// Test getting ticket
	ticket, err := store.GetTicketByEmail(user.Email)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ticket.User.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, ticket.User.Email)
	}

	// Test getting non-existent ticket
	_, err = store.GetTicketByEmail("nonexistent@example.com")
	if err != ErrTicketNotFound {
		t.Errorf("Expected ErrTicketNotFound, got: %v", err)
	}
}

func TestGetAllAllocations(t *testing.T) {
	store := NewStore()

	// Purchase tickets in different sections
	user1 := model.User{Email: "user1@example.com", FirstName: "User1", LastName: "One"}
	user2 := model.User{Email: "user2@example.com", FirstName: "User2", LastName: "Two"}

	store.PurchaseTicket(user1, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	store.PurchaseTicket(user2, config.RouteFrom, config.RouteTo, config.TicketPriceCents)

	// Test get all allocations
	allocations := store.GetAllAllocations("")
	if len(allocations) != 2 {
		t.Errorf("Expected 2 allocations, got %d", len(allocations))
	}

	// Test filter by section
	allocationsA := store.GetAllAllocations("A")
	if len(allocationsA) > 2 {
		t.Errorf("Expected at most 2 allocations in section A, got %d", len(allocationsA))
	}
}

func TestRemoveTicket(t *testing.T) {
	store := NewStore()

	user := model.User{
		Email:     "remove@example.com",
		FirstName: "Remove",
		LastName:  "Me",
	}

	// Purchase ticket
	_, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	// Verify ticket exists
	_, err = store.GetTicketByEmail(user.Email)
	if err != nil {
		t.Fatalf("Ticket should exist: %v", err)
	}

	// Remove ticket
	err = store.RemoveTicket(user.Email)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify ticket is removed
	_, err = store.GetTicketByEmail(user.Email)
	if err != ErrTicketNotFound {
		t.Errorf("Expected ErrTicketNotFound, got: %v", err)
	}
}

func TestModifySeat(t *testing.T) {
	store := NewStore()

	user := model.User{
		Email:     "modify@example.com",
		FirstName: "Modify",
		LastName:  "Seat",
	}

	// Purchase ticket
	ticket, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket: %v", err)
	}

	originalSection := ticket.Seat.Section

	// Determine new section (opposite of current)
	newSection := "B"
	if originalSection == "B" {
		newSection = "A"
	}
	newSeatNumber := int32(5)

	// Modify seat
	updatedTicket, err := store.ModifySeat(user.Email, newSection, newSeatNumber)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updatedTicket.Seat.Section != newSection {
		t.Errorf("Expected section %s, got %s", newSection, updatedTicket.Seat.Section)
	}

	if updatedTicket.Seat.SeatNumber != newSeatNumber {
		t.Errorf("Expected seat number %d, got %d", newSeatNumber, updatedTicket.Seat.SeatNumber)
	}

	// Test invalid section
	_, err = store.ModifySeat(user.Email, "C", 1)
	if err == nil {
		t.Error("Expected error for invalid section")
	}

	// Test invalid seat number
	_, err = store.ModifySeat(user.Email, "A", 11)
	if err == nil {
		t.Error("Expected error for invalid seat number")
	}
}

func TestSeatAllocationOrder(t *testing.T) {
	store := NewStore()

	// Purchase 10 tickets - should fill section A
	for i := 1; i <= 10; i++ {
		user := model.User{
			Email:     fmt.Sprintf("user%d@example.com", i),
			FirstName: "User",
			LastName:  fmt.Sprintf("%d", i),
		}
		ticket, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
		if err != nil {
			t.Fatalf("Failed to purchase ticket %d: %v", i, err)
		}

		// First 10 should be in section A
		if i <= 10 && ticket.Seat.Section != "A" {
			t.Errorf("Ticket %d: Expected section A, got %s", i, ticket.Seat.Section)
		}

		// Seat numbers should be 1-10
		if ticket.Seat.SeatNumber != int32(i) {
			t.Errorf("Ticket %d: Expected seat number %d, got %d", i, i, ticket.Seat.SeatNumber)
		}
	}

	// Next ticket should be in section B
	user11 := model.User{Email: "user11@example.com", FirstName: "User", LastName: "11"}
	ticket11, err := store.PurchaseTicket(user11, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		t.Fatalf("Failed to purchase ticket 11: %v", err)
	}

	if ticket11.Seat.Section != "B" {
		t.Errorf("Ticket 11: Expected section B, got %s", ticket11.Seat.Section)
	}

	if ticket11.Seat.SeatNumber != 1 {
		t.Errorf("Ticket 11: Expected seat number 1, got %d", ticket11.Seat.SeatNumber)
	}
}

func TestTrainFull(t *testing.T) {
	store := NewStore()

	// Fill all 20 seats
	for i := 1; i <= 20; i++ {
		user := model.User{
			Email:     fmt.Sprintf("user%d@example.com", i),
			FirstName: "User",
			LastName:  fmt.Sprintf("%d", i),
		}
		_, err := store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
		if err != nil {
			t.Fatalf("Failed to purchase ticket %d: %v", i, err)
		}
	}

	// Try to purchase one more - should fail
	user21 := model.User{Email: "user21@example.com", FirstName: "User", LastName: "21"}
	_, err := store.PurchaseTicket(user21, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != ErrTrainFull {
		t.Errorf("Expected ErrTrainFull, got: %v", err)
	}
}

