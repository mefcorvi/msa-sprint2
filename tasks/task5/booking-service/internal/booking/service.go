package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// Monolith reads data of the modules that are still in the monolith. Lookups return nil (or false) when
// the monolith does not know the entity, and an error wrapping ErrMonolithUnavailable when it cannot answer.
type Monolith interface {
	FindUser(ctx context.Context, userID string) (*User, error)
	FindHotel(ctx context.Context, hotelID string) (*Hotel, error)
	IsHotelTrusted(ctx context.Context, hotelID string) (bool, error)
	FindPromoDiscount(ctx context.Context, code, userID string) (discount float64, found bool, err error)
}

// Store persists a booking and the event built from the saved booking (with its id) in one transaction.
type Store interface {
	Save(ctx context.Context, draft Booking, newEvent func(saved Booking) CreatedEvent) (Booking, error)
	List(ctx context.Context, userID string) ([]Booking, error)
}

// Service ports the booking rules from the monolith one-to-one, questionable ones included (ADR-002, step 1):
// the regression suite must behave identically against both implementations before cutover.
type Service struct {
	monolith Monolith
	store    Store
	now      func() time.Time
	log      zerolog.Logger
}

func NewService(monolith Monolith, store Store, now func() time.Time, log zerolog.Logger) *Service {
	return &Service{monolith: monolith, store: store, now: now, log: log}
}

// List returns all bookings when userID is empty.
func (s *Service) List(ctx context.Context, userID string) ([]Booking, error) {
	return s.store.List(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID, hotelID, promoCode string) (Booking, error) {
	s.log.Info().Str("userId", userID).Str("hotelId", hotelID).Str("promoCode", promoCode).Msg("creating booking")

	user, err := s.monolith.FindUser(ctx, userID)
	if err != nil {
		return Booking{}, err
	}
	if err := s.validateUser(userID, user); err != nil {
		return Booking{}, err
	}
	if err := s.validateHotel(ctx, hotelID); err != nil {
		return Booking{}, err
	}

	basePrice := BasePrice
	if user.IsVIP() {
		basePrice = VIPBasePrice
	}
	discount, err := s.promoDiscount(ctx, promoCode, userID)
	if err != nil {
		return Booking{}, err
	}
	// The discount is subtracted as an absolute amount, as in the monolith.
	price := basePrice - discount
	s.log.Info().Float64("base", basePrice).Float64("discount", discount).Float64("final", price).Msg("final price calculated")

	return s.store.Save(ctx, Booking{
		UserID:          userID,
		HotelID:         hotelID,
		PromoCode:       promoCode,
		DiscountPercent: discount,
		Price:           price,
		// PostgreSQL keeps microseconds; truncating keeps the response equal to what is persisted.
		CreatedAt: s.now().UTC().Truncate(time.Microsecond),
	}, func(saved Booking) CreatedEvent {
		return NewCreatedEvent(saved, s.now().UTC())
	})
}

func (s *Service) validateUser(userID string, user *User) error {
	if user == nil || !user.Active {
		return s.reject("User %s is inactive", userID)
	}
	if user.Blacklisted {
		return s.reject("User %s is blacklisted", userID)
	}
	return nil
}

func (s *Service) validateHotel(ctx context.Context, hotelID string) error {
	hotel, err := s.monolith.FindHotel(ctx, hotelID)
	if err != nil {
		return err
	}
	if hotel == nil || !hotel.Operational {
		return s.reject("Hotel %s is not operational", hotelID)
	}
	trusted, err := s.monolith.IsHotelTrusted(ctx, hotelID)
	if err != nil {
		return err
	}
	if !trusted {
		return s.reject("Hotel %s is not trusted based on reviews", hotelID)
	}
	if hotel.FullyBooked {
		return s.reject("Hotel %s is fully booked", hotelID)
	}
	return nil
}

func (s *Service) promoDiscount(ctx context.Context, promoCode, userID string) (float64, error) {
	if promoCode == "" {
		return 0, nil
	}
	discount, found, err := s.monolith.FindPromoDiscount(ctx, promoCode, userID)
	if err != nil {
		return 0, err
	}
	if !found {
		s.log.Info().Str("promoCode", promoCode).Str("userId", userID).Msg("promo code is invalid or not applicable")
		return 0, nil
	}
	return discount, nil
}

func (s *Service) reject(format string, args ...any) error {
	reason := fmt.Sprintf(format, args...)
	s.log.Warn().Str("reason", reason).Msg("booking rejected")
	return &RejectedError{Reason: reason}
}
