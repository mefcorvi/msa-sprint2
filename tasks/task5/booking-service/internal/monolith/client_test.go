package monolith

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotelio/booking-service/internal/booking"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, time.Second)
}

func TestFindUserTranslatesMonolithPayload(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users/u1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"id":"u1","status":"VIP","blacklisted":false,"active":true,"email":"u1@test.com"}`))
	})

	user, err := c.FindUser(context.Background(), "u1")

	if err != nil || user == nil || !user.IsVIP() || !user.Active || user.Blacklisted {
		t.Fatalf("user = %+v, err = %v", user, err)
	}
}

func TestNotFoundMeansNoEntity(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	user, err := c.FindUser(context.Background(), "ghost")

	if err != nil || user != nil {
		t.Fatalf("user = %+v, err = %v", user, err)
	}
}

func TestRejectedPromoIsNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Query().Get("code") != "OLD" || r.URL.Query().Get("userId") != "u1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	_, found, err := c.FindPromoDiscount(context.Background(), "OLD", "u1")

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}

func TestServerErrorMeansMonolithUnavailable(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := c.IsHotelTrusted(context.Background(), "h1")

	if !errors.Is(err, booking.ErrMonolithUnavailable) {
		t.Fatalf("err = %v, want ErrMonolithUnavailable", err)
	}
}
