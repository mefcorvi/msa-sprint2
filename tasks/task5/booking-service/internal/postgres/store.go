package postgres

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"hotelio/booking-service/internal/booking"
	"hotelio/booking-service/internal/outbox"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Save(ctx context.Context, b booking.Booking, newEvent func(booking.Booking) booking.CreatedEvent) (booking.Booking, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO booking (user_id, hotel_id, promo_code, discount_percent, price, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`,
			b.UserID, b.HotelID, nullIfEmpty(b.PromoCode), b.DiscountPercent, b.Price, b.CreatedAt,
		).Scan(&b.ID)
		if err != nil {
			return err
		}
		event := newEvent(b)
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}
		return outbox.Add(ctx, tx, event.EventID, event.Type, strconv.FormatInt(b.ID, 10), payload)
	})
	return b, err
}

// List returns all bookings when userID is empty.
func (s *Store) List(ctx context.Context, userID string) ([]booking.Booking, error) {
	query := `SELECT id, user_id, hotel_id, promo_code, discount_percent, price, created_at FROM booking`
	var args []any
	if userID != "" {
		query += ` WHERE user_id = $1`
		args = append(args, userID)
	}
	rows, err := s.pool.Query(ctx, query+` ORDER BY id`, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (booking.Booking, error) {
		var b booking.Booking
		var promoCode *string
		err := row.Scan(&b.ID, &b.UserID, &b.HotelID, &promoCode, &b.DiscountPercent, &b.Price, &b.CreatedAt)
		if promoCode != nil {
			b.PromoCode = *promoCode
		}
		return b, err
	})
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
