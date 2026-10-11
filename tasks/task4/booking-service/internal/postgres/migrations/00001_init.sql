-- +goose Up
CREATE TABLE booking (
    id               BIGSERIAL PRIMARY KEY,
    user_id          VARCHAR(255)     NOT NULL,
    hotel_id         VARCHAR(255)     NOT NULL,
    promo_code       VARCHAR(255),
    discount_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    price            DOUBLE PRECISION NOT NULL,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX idx_booking_user_id ON booking (user_id);

-- Transactional outbox: events are written in the same transaction as the booking
-- and published to Kafka by the outbox relay.
CREATE TABLE outbox_event (
    id           BIGSERIAL PRIMARY KEY,
    event_id     UUID         NOT NULL UNIQUE,
    event_type   VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    payload      TEXT         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_event_unpublished ON outbox_event (id) WHERE published_at IS NULL;

-- +goose Down
DROP TABLE outbox_event;
DROP TABLE booking;
