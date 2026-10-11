package booking

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	BasePrice    = 100.0
	VIPBasePrice = 80.0
)

// ErrMonolithUnavailable is returned when data of a module that still lives in the monolith cannot be read.
var ErrMonolithUnavailable = errors.New("monolith unavailable")

// Booking has no promo code when PromoCode is empty.
type Booking struct {
	ID              int64
	UserID          string
	HotelID         string
	PromoCode       string
	DiscountPercent float64
	Price           float64
	CreatedAt       time.Time
}

// User and Hotel are this service's view of monolith data; the anti-corruption layer maps into them.
type User struct {
	Status      string
	Active      bool
	Blacklisted bool
}

func (u User) IsVIP() bool {
	return strings.EqualFold(u.Status, "VIP")
}

type Hotel struct {
	Operational bool
	FullyBooked bool
}

type RejectedError struct {
	Reason string
}

func (e *RejectedError) Error() string {
	return e.Reason
}

const CreatedEventType = "BookingCreated"

type CreatedEvent struct {
	EventID         string    `json:"eventId"`
	Type            string    `json:"type"`
	OccurredAt      time.Time `json:"occurredAt"`
	BookingID       int64     `json:"bookingId"`
	UserID          string    `json:"userId"`
	HotelID         string    `json:"hotelId"`
	PromoCode       *string   `json:"promoCode"`
	DiscountPercent float64   `json:"discountPercent"`
	Price           float64   `json:"price"`
	CreatedAt       time.Time `json:"createdAt"`
}

func NewCreatedEvent(b Booking, occurredAt time.Time) CreatedEvent {
	var promoCode *string
	if b.PromoCode != "" {
		promoCode = &b.PromoCode
	}
	return CreatedEvent{
		EventID:         uuid.NewString(),
		Type:            CreatedEventType,
		OccurredAt:      occurredAt,
		BookingID:       b.ID,
		UserID:          b.UserID,
		HotelID:         b.HotelID,
		PromoCode:       promoCode,
		DiscountPercent: b.DiscountPercent,
		Price:           b.Price,
		CreatedAt:       b.CreatedAt,
	}
}
