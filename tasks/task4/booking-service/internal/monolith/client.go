// Package monolith is the anti-corruption layer between booking-service and the monolith: it reads the
// monolith's REST API and translates its payloads and failures into the booking domain model.
package monolith

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"hotelio/booking-service/internal/booking"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type userDTO struct {
	Status      string `json:"status"`
	Active      bool   `json:"active"`
	Blacklisted bool   `json:"blacklisted"`
}

type hotelDTO struct {
	Operational bool `json:"operational"`
	FullyBooked bool `json:"fullyBooked"`
}

type promoDTO struct {
	Discount float64 `json:"discount"`
}

func (c *Client) FindUser(ctx context.Context, userID string) (*booking.User, error) {
	var dto userDTO
	found, err := c.call(ctx, http.MethodGet, "/api/users/"+url.PathEscape(userID), nil, &dto)
	if err != nil || !found {
		return nil, err
	}
	return &booking.User{Status: dto.Status, Active: dto.Active, Blacklisted: dto.Blacklisted}, nil
}

func (c *Client) FindHotel(ctx context.Context, hotelID string) (*booking.Hotel, error) {
	var dto hotelDTO
	found, err := c.call(ctx, http.MethodGet, "/api/hotels/"+url.PathEscape(hotelID), nil, &dto)
	if err != nil || !found {
		return nil, err
	}
	return &booking.Hotel{Operational: dto.Operational, FullyBooked: dto.FullyBooked}, nil
}

func (c *Client) IsHotelTrusted(ctx context.Context, hotelID string) (bool, error) {
	var trusted bool
	_, err := c.call(ctx, http.MethodGet, "/api/reviews/hotel/"+url.PathEscape(hotelID)+"/trusted", nil, &trusted)
	return trusted, err
}

// FindPromoDiscount reports found=false for unknown, expired or not applicable codes (the monolith answers 400).
func (c *Client) FindPromoDiscount(ctx context.Context, code, userID string) (float64, bool, error) {
	var dto promoDTO
	query := url.Values{"code": {code}, "userId": {userID}}
	found, err := c.call(ctx, http.MethodPost, "/api/promos/validate", query, &dto)
	return dto.Discount, found, err
}

// call decodes a 2xx response into out; 404 and 400 mean "no such entity" in the monolith API.
func (c *Client) call(ctx context.Context, method, path string, query url.Values, out any) (bool, error) {
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, target, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %s %s: %v", booking.ErrMonolithUnavailable, method, path, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest:
		return false, nil
	case resp.StatusCode >= 300:
		return false, fmt.Errorf("%w: %s %s: HTTP %d", booking.ErrMonolithUnavailable, method, path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return false, fmt.Errorf("%w: %s %s: decode response: %v", booking.ErrMonolithUnavailable, method, path, err)
	}
	return true, nil
}
