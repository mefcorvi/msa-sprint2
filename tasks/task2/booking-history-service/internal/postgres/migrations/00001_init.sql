-- +goose Up
CREATE TABLE booking_history (
    booking_id       BIGINT PRIMARY KEY,
    event_id         UUID             NOT NULL,
    user_id          VARCHAR(255)     NOT NULL,
    hotel_id         VARCHAR(255)     NOT NULL,
    promo_code       VARCHAR(255),
    discount_percent DOUBLE PRECISION NOT NULL,
    price            DOUBLE PRECISION NOT NULL,
    created_at       TIMESTAMPTZ      NOT NULL,
    received_at      TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX idx_booking_history_user_id ON booking_history (user_id);
CREATE INDEX idx_booking_history_hotel_id ON booking_history (hotel_id);
CREATE INDEX idx_booking_history_created_at ON booking_history (created_at);

CREATE VIEW booking_stats_by_day AS
SELECT (created_at AT TIME ZONE 'UTC')::date                AS day,
       count(*)                                         AS bookings,
       sum(price)                                       AS revenue,
       count(*) FILTER (WHERE promo_code IS NOT NULL)   AS with_promo
FROM booking_history
GROUP BY 1;

CREATE VIEW booking_stats_by_hotel AS
SELECT hotel_id,
       count(*)                AS bookings,
       count(DISTINCT user_id) AS unique_users,
       sum(price)              AS revenue,
       avg(price)              AS avg_price
FROM booking_history
GROUP BY hotel_id;

CREATE VIEW booking_stats_by_user AS
SELECT user_id,
       count(*)                 AS bookings,
       count(DISTINCT hotel_id) AS hotels,
       sum(price)               AS spent,
       max(created_at)          AS last_booking_at
FROM booking_history
GROUP BY user_id;

-- +goose Down
DROP VIEW booking_stats_by_user;
DROP VIEW booking_stats_by_hotel;
DROP VIEW booking_stats_by_day;
DROP TABLE booking_history;
