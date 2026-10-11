package history

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type fakeStore struct {
	saved []BookingCreated
}

func (s *fakeStore) Save(_ context.Context, e BookingCreated) (bool, error) {
	s.saved = append(s.saved, e)
	return true, nil
}

func newHandler() (*Handler, *fakeStore) {
	store := &fakeStore{}
	return NewHandler(store, zerolog.Nop()), store
}

func TestRecordsBookingCreatedEvent(t *testing.T) {
	h, store := newHandler()

	err := h.Handle(context.Background(), []byte(`{"eventId":"0b7c7a8e-6f2a-4a39-9d55-1f7c1b7b1a01","type":"BookingCreated",
		"occurredAt":"2026-10-10T12:00:01Z","bookingId":7,"userId":"u1","hotelId":"h1",
		"promoCode":null,"discountPercent":0,"price":100,"createdAt":"2026-10-10T12:00:00Z"}`))

	if err != nil {
		t.Fatal(err)
	}
	if len(store.saved) != 1 {
		t.Fatalf("saved %d events, want 1", len(store.saved))
	}
	e := store.saved[0]
	if e.BookingID != 7 || e.PromoCode != nil || !e.CreatedAt.Equal(time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected event %+v", e)
	}
}

func TestIgnoresOtherEventTypes(t *testing.T) {
	h, store := newHandler()

	err := h.Handle(context.Background(), []byte(`{"eventId":"x","type":"BookingCancelled","bookingId":7}`))

	if err != nil || len(store.saved) != 0 {
		t.Fatalf("err = %v, saved = %d", err, len(store.saved))
	}
}

func TestMalformedPayloadIsInvalid(t *testing.T) {
	h, _ := newHandler()

	err := h.Handle(context.Background(), []byte(`not json`))

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("err = %v, want ErrInvalidEvent", err)
	}
}
