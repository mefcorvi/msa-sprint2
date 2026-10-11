package grpcapi

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"hotelio/booking-service/internal/booking"
	"hotelio/booking-service/internal/pb"
)

type BookingService interface {
	Create(ctx context.Context, userID, hotelID, promoCode string) (booking.Booking, error)
	List(ctx context.Context, userID string) ([]booking.Booking, error)
}

type Server struct {
	pb.UnimplementedBookingServiceServer
	bookings BookingService
	log      zerolog.Logger
}

func NewServer(bookings BookingService, log zerolog.Logger) *Server {
	return &Server{bookings: bookings, log: log}
}

// proto3 strings cannot be null: an absent promo code or user id arrives as "" and is returned as "".

func (s *Server) CreateBooking(ctx context.Context, req *pb.BookingRequest) (*pb.BookingResponse, error) {
	b, err := s.bookings.Create(ctx, req.GetUserId(), req.GetHotelId(), req.GetPromoCode())
	if err != nil {
		return nil, s.toStatus(err)
	}
	return toResponse(b), nil
}

func (s *Server) ListBookings(ctx context.Context, req *pb.BookingListRequest) (*pb.BookingListResponse, error) {
	bookings, err := s.bookings.List(ctx, req.GetUserId())
	if err != nil {
		return nil, s.toStatus(err)
	}
	resp := &pb.BookingListResponse{Bookings: make([]*pb.BookingResponse, len(bookings))}
	for i, b := range bookings {
		resp.Bookings[i] = toResponse(b)
	}
	return resp, nil
}

func (s *Server) toStatus(err error) error {
	var rejected *booking.RejectedError
	switch {
	case errors.As(err, &rejected):
		return status.Error(codes.FailedPrecondition, rejected.Reason)
	case errors.Is(err, booking.ErrMonolithUnavailable):
		s.log.Error().Err(err).Msg("monolith call failed")
		return status.Error(codes.Unavailable, err.Error())
	default:
		s.log.Error().Err(err).Msg("unexpected error")
		return status.Error(codes.Internal, err.Error())
	}
}

func toResponse(b booking.Booking) *pb.BookingResponse {
	return &pb.BookingResponse{
		Id:              strconv.FormatInt(b.ID, 10),
		UserId:          b.UserID,
		HotelId:         b.HotelID,
		PromoCode:       b.PromoCode,
		DiscountPercent: b.DiscountPercent,
		Price:           b.Price,
		CreatedAt:       b.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
