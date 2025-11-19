package service

import (
	"context"

	ticket "github.com/cloudbees/train-ticket-service/api"
	"github.com/cloudbees/train-ticket-service/internal/auth"
	"github.com/cloudbees/train-ticket-service/internal/config"
	"github.com/cloudbees/train-ticket-service/internal/model"
	"github.com/cloudbees/train-ticket-service/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TicketService struct {
	ticket.UnimplementedTicketServiceServer
	store *store.Store
}

func NewTicketService(s *store.Store) *TicketService {
	return &TicketService{
		store: s,
	}
}

func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (*ticket.PurchaseTicketResponse, error) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "first_name, last_name, and email are required")
	}

	user := model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	t, err := s.store.PurchaseTicket(user, config.RouteFrom, config.RouteTo, config.TicketPriceCents)
	if err != nil {
		switch err {
		case store.ErrUserAlreadyHasTicket:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case store.ErrTrainFull:
			return nil, status.Error(codes.ResourceExhausted, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &ticket.PurchaseTicketResponse{
		Receipt: convertTicketToReceipt(t),
	}, nil
}

func (s *TicketService) ViewUserReceipt(ctx context.Context, req *ticket.ViewUserReceiptRequest) (*ticket.ViewUserReceiptResponse, error) {
	userClaims, err := auth.ExtractUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	t, err := s.store.GetTicketByEmail(userClaims.Email)
	if err != nil {
		if err == store.ErrTicketNotFound {
			return nil, status.Error(codes.NotFound, "ticket not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ticket.ViewUserReceiptResponse{
		Receipt: convertTicketToReceipt(t),
	}, nil
}

func (s *TicketService) ViewAllocations(ctx context.Context, req *ticket.ViewAllocationsRequest) (*ticket.ViewAllocationsResponse, error) {
	userClaims, err := auth.ExtractUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if !userClaims.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	allocations := s.store.GetAllAllocations(req.Section)

	protoAllocations := make([]*ticket.Allocation, 0, len(allocations))
	for _, t := range allocations {
		protoAllocations = append(protoAllocations, &ticket.Allocation{
			Section:    t.Seat.Section,
			SeatNumber: t.Seat.SeatNumber,
			User: &ticket.User{
				FirstName: t.User.FirstName,
				LastName:  t.User.LastName,
				Email:     t.User.Email,
			},
		})
	}

	return &ticket.ViewAllocationsResponse{
		Allocations: protoAllocations,
	}, nil
}

func (s *TicketService) RemoveUserFromTrain(ctx context.Context, req *ticket.RemoveUserFromTrainRequest) (*ticket.RemoveUserFromTrainResponse, error) {
	userClaims, err := auth.ExtractUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	targetEmail := userClaims.Email
	if req.Email != "" {
		if !userClaims.IsAdmin() {
			return nil, status.Error(codes.PermissionDenied, "only admin can remove other users")
		}
		targetEmail = req.Email
	}

	err = s.store.RemoveTicket(targetEmail)
	if err != nil {
		if err == store.ErrTicketNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ticket.RemoveUserFromTrainResponse{
		Success: true,
		Message: "user removed from train successfully",
	}, nil
}

func (s *TicketService) ModifyUserSeat(ctx context.Context, req *ticket.ModifyUserSeatRequest) (*ticket.ModifyUserSeatResponse, error) {
	userClaims, err := auth.ExtractUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if req.Section == "" || req.SeatNumber == 0 {
		return nil, status.Error(codes.InvalidArgument, "section and seat_number are required")
	}

	targetEmail := userClaims.Email
	if req.Email != "" {
		if !userClaims.IsAdmin() {
			return nil, status.Error(codes.PermissionDenied, "only admin can modify other users' seats")
		}
		targetEmail = req.Email
	}

	t, err := s.store.ModifySeat(targetEmail, req.Section, req.SeatNumber)
	if err != nil {
		switch err {
		case store.ErrTicketNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case store.ErrSeatAlreadyOccupied:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case store.ErrInvalidSeat:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &ticket.ModifyUserSeatResponse{
		Receipt: convertTicketToReceipt(t),
	}, nil
}

func convertTicketToReceipt(t *model.Ticket) *ticket.Receipt {
	return &ticket.Receipt{
		From: t.From,
		To:   t.To,
		User: &ticket.User{
			FirstName: t.User.FirstName,
			LastName:  t.User.LastName,
			Email:     t.User.Email,
		},
		PricePaid: t.PricePaid,
		Seat: &ticket.Seat{
			Section:    t.Seat.Section,
			SeatNumber: t.Seat.SeatNumber,
		},
	}
}

