package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cloudbees/train-ticket-service/internal/model"
)

var (
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrSeatAlreadyOccupied  = errors.New("seat is already occupied")
	ErrTrainFull            = errors.New("train is full")
	ErrInvalidSeat          = errors.New("invalid seat")
	ErrUserAlreadyHasTicket = errors.New("user already has a ticket")
)

type Store struct {
	mu      sync.RWMutex
	tickets map[string]*model.Ticket 
	seats   map[string]bool          
}

func NewStore() *Store {
	return &Store{
		tickets: make(map[string]*model.Ticket),
		seats:   make(map[string]bool),
	}
}

func (s *Store) PurchaseTicket(user model.User, from, to string, pricePaid int32) (*model.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tickets[user.Email]; exists {
		return nil, ErrUserAlreadyHasTicket
	}

	seat, err := s.findNextAvailableSeat()
	if err != nil {
		return nil, err
	}

	ticket := &model.Ticket{
		From:      from,
		To:        to,
		User:      user,
		PricePaid: pricePaid,
		Seat:      *seat,
	}

	s.tickets[user.Email] = ticket
	s.seats[seatKey(seat.Section, seat.SeatNumber)] = true

	return ticket, nil
}

func (s *Store) GetTicketByEmail(email string) (*model.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, exists := s.tickets[email]
	if !exists {
		return nil, ErrTicketNotFound
	}

	return ticket, nil
}

func (s *Store) GetAllAllocations(sectionFilter string) []*model.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allocations []*model.Ticket
	for _, ticket := range s.tickets {
		if sectionFilter == "" || ticket.Seat.Section == sectionFilter {
			allocations = append(allocations, ticket)
		}
	}

	return allocations
}

func (s *Store) RemoveTicket(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[email]
	if !exists {
		return ErrTicketNotFound
	}

	delete(s.seats, seatKey(ticket.Seat.Section, ticket.Seat.SeatNumber))

	delete(s.tickets, email)

	return nil
}

func (s *Store) ModifySeat(email, newSection string, newSeatNumber int32) (*model.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !model.IsValidSection(newSection) {
		return nil, fmt.Errorf("%w: invalid section %s", ErrInvalidSeat, newSection)
	}
	if !model.IsValidSeatNumber(newSeatNumber) {
		return nil, fmt.Errorf("%w: invalid seat number %d", ErrInvalidSeat, newSeatNumber)
	}

	ticket, exists := s.tickets[email]
	if !exists {
		return nil, ErrTicketNotFound
	}

	newSeatKey := seatKey(newSection, newSeatNumber)
	if s.seats[newSeatKey] {
		return nil, ErrSeatAlreadyOccupied
	}

	oldSeatKey := seatKey(ticket.Seat.Section, ticket.Seat.SeatNumber)
	delete(s.seats, oldSeatKey)

	ticket.Seat.Section = newSection
	ticket.Seat.SeatNumber = newSeatNumber
	s.seats[newSeatKey] = true

	return ticket, nil
}

func (s *Store) findNextAvailableSeat() (*model.Seat, error) {
	for i := int32(1); i <= 10; i++ {
		key := seatKey("A", i)
		if !s.seats[key] {
			return &model.Seat{Section: "A", SeatNumber: i}, nil
		}
	}

	for i := int32(1); i <= 10; i++ {
		key := seatKey("B", i)
		if !s.seats[key] {
			return &model.Seat{Section: "B", SeatNumber: i}, nil
		}
	}

	return nil, ErrTrainFull
}

func seatKey(section string, seatNumber int32) string {
	return fmt.Sprintf("%s-%d", section, seatNumber)
}

