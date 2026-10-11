package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

var now = time.Date(2026, 10, 10, 12, 0, 0, 123456789, time.UTC)

type fakeMonolith struct {
	users   map[string]User
	hotels  map[string]Hotel
	trusted map[string]bool
	promos  map[string]float64
}

func (m *fakeMonolith) FindUser(_ context.Context, id string) (*User, error) {
	if u, ok := m.users[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (m *fakeMonolith) FindHotel(_ context.Context, id string) (*Hotel, error) {
	if h, ok := m.hotels[id]; ok {
		return &h, nil
	}
	return nil, nil
}

func (m *fakeMonolith) IsHotelTrusted(_ context.Context, id string) (bool, error) {
	return m.trusted[id], nil
}

func (m *fakeMonolith) FindPromoDiscount(_ context.Context, code, _ string) (float64, bool, error) {
	d, ok := m.promos[code]
	return d, ok, nil
}

type fakeStore struct {
	saved  []Booking
	events []CreatedEvent
}

func (s *fakeStore) Save(_ context.Context, draft Booking, newEvent func(Booking) CreatedEvent) (Booking, error) {
	draft.ID = 42
	s.saved = append(s.saved, draft)
	s.events = append(s.events, newEvent(draft))
	return draft, nil
}

func (s *fakeStore) List(context.Context, string) ([]Booking, error) {
	return s.saved, nil
}

func newService() (*Service, *fakeMonolith, *fakeStore) {
	m := &fakeMonolith{
		users: map[string]User{
			"regular":     {Status: "ACTIVE", Active: true},
			"vip":         {Status: "VIP", Active: true},
			"inactive":    {Status: "ACTIVE"},
			"blacklisted": {Status: "ACTIVE", Active: true, Blacklisted: true},
		},
		hotels: map[string]Hotel{
			"open":   {Operational: true},
			"closed": {},
			"full":   {Operational: true, FullyBooked: true},
		},
		trusted: map[string]bool{"open": true, "closed": true, "full": true},
		promos:  map[string]float64{"VIP20": 20},
	}
	store := &fakeStore{}
	return NewService(m, store, func() time.Time { return now }, zerolog.Nop()), m, store
}

func TestRegularUserWithoutPromoPaysBasePrice(t *testing.T) {
	svc, _, _ := newService()

	b, err := svc.Create(context.Background(), "regular", "open", "")

	if err != nil {
		t.Fatal(err)
	}
	if b.ID != 42 || b.Price != 100 || b.DiscountPercent != 0 {
		t.Fatalf("unexpected booking %+v", b)
	}
	if want := now.Truncate(time.Microsecond); !b.CreatedAt.Equal(want) {
		t.Fatalf("createdAt = %v, want %v", b.CreatedAt, want)
	}
}

func TestVIPUserWithPromoKeepsMonolithPricing(t *testing.T) {
	svc, _, _ := newService()

	b, err := svc.Create(context.Background(), "vip", "open", "VIP20")

	if err != nil {
		t.Fatal(err)
	}
	if b.Price != 60 || b.DiscountPercent != 20 {
		t.Fatalf("unexpected booking %+v", b)
	}
}

func TestUnknownPromoGivesNoDiscount(t *testing.T) {
	svc, _, _ := newService()

	b, err := svc.Create(context.Background(), "regular", "open", "NOPE")

	if err != nil {
		t.Fatal(err)
	}
	if b.Price != 100 {
		t.Fatalf("price = %v, want 100", b.Price)
	}
}

func TestRejections(t *testing.T) {
	tests := []struct {
		user, hotel, reason string
	}{
		{"ghost", "open", "User ghost is inactive"},
		{"inactive", "open", "User inactive is inactive"},
		{"blacklisted", "open", "User blacklisted is blacklisted"},
		{"regular", "closed", "Hotel closed is not operational"},
		{"regular", "unknown", "Hotel unknown is not operational"},
		{"regular", "full", "Hotel full is fully booked"},
	}
	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			svc, _, store := newService()

			_, err := svc.Create(context.Background(), tt.user, tt.hotel, "")

			var rejected *RejectedError
			if !errors.As(err, &rejected) || rejected.Reason != tt.reason {
				t.Fatalf("err = %v, want rejection %q", err, tt.reason)
			}
			if len(store.saved) != 0 {
				t.Fatal("rejected booking must not be saved")
			}
		})
	}
}

func TestUntrustedHotelIsRejected(t *testing.T) {
	svc, m, _ := newService()
	m.trusted["open"] = false

	_, err := svc.Create(context.Background(), "regular", "open", "")

	var rejected *RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != "Hotel open is not trusted based on reviews" {
		t.Fatalf("err = %v", err)
	}
}

func TestCreatedBookingIsSavedWithItsEvent(t *testing.T) {
	svc, _, store := newService()

	b, err := svc.Create(context.Background(), "regular", "open", "")

	if err != nil {
		t.Fatal(err)
	}
	if len(store.events) != 1 {
		t.Fatalf("saved %d events, want 1", len(store.events))
	}
	e := store.events[0]
	if e.Type != CreatedEventType || e.BookingID != b.ID || e.UserID != "regular" || e.PromoCode != nil || e.EventID == "" || !e.OccurredAt.Equal(now) {
		t.Fatalf("unexpected event %+v", e)
	}
}
