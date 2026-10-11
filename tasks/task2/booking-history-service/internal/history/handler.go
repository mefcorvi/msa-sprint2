package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

const BookingCreatedType = "BookingCreated"

// ErrInvalidEvent marks payloads that will never be processed, so retrying them is pointless.
var ErrInvalidEvent = errors.New("invalid event")

type BookingCreated struct {
	EventID         string    `json:"eventId"`
	Type            string    `json:"type"`
	BookingID       int64     `json:"bookingId"`
	UserID          string    `json:"userId"`
	HotelID         string    `json:"hotelId"`
	PromoCode       *string   `json:"promoCode"`
	DiscountPercent float64   `json:"discountPercent"`
	Price           float64   `json:"price"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Store interface {
	// Save reports false when the booking is already recorded (a redelivered event).
	Save(ctx context.Context, e BookingCreated) (bool, error)
}

type Handler struct {
	store Store
	log   zerolog.Logger
}

func NewHandler(store Store, log zerolog.Logger) *Handler {
	return &Handler{store: store, log: log}
}

func (h *Handler) Handle(ctx context.Context, payload []byte) error {
	var e BookingCreated
	if err := json.Unmarshal(payload, &e); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEvent, err)
	}
	if e.Type != BookingCreatedType {
		h.log.Debug().Str("type", e.Type).Msg("skipping event")
		return nil
	}
	recorded, err := h.store.Save(ctx, e)
	if err != nil {
		return err
	}
	if recorded {
		h.log.Info().Int64("bookingId", e.BookingID).Str("userId", e.UserID).Str("hotelId", e.HotelID).Float64("price", e.Price).Msg("recorded booking")
	} else {
		h.log.Info().Int64("bookingId", e.BookingID).Str("eventId", e.EventID).Msg("booking already recorded, duplicate event ignored")
	}
	return nil
}
