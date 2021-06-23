-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

CREATE TYPE payment_status AS ENUM ('pending', 'failed', 'completed');
CREATE TYPE purchase_status AS ENUM ('pending', 'failed', 'completed');

CREATE TABLE purchases (
   id                VARCHAR(24)       PRIMARY KEY,
   buyer_id          VARCHAR(24)       NOT NULL, -- FOREIGN KEY REFERENCES vb_users/users(id)
   items             JSONB             NOT NULL,
   status            purchase_status   NOT NULL DEFAULT 'pending',
   created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id                VARCHAR(24)       PRIMARY KEY,
    origin            VARCHAR(24)       NOT NULL REFERENCES PURCHASES(id) ON DELETE CASCADE,
    buyer_id          VARCHAR(24)       NOT NULL,
    seller            VARCHAR(24)       NOT NULL,
    payment_status    payment_status    NOT NULL DEFAULT 'pending',
    items             JSONB             NOT NULL,
    other_status      JSONB             NULL,
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE TABLE bookings (
    id                VARCHAR(24)       PRIMARY KEY,
    origin            VARCHAR(24)       NOT NULL REFERENCES PURCHASES(id) ON DELETE CASCADE,
    buyer_id          VARCHAR(24)       NOT NULL,
    host              VARCHAR(24)       NOT NULL,
    item_id           VARCHAR(24)       NOT NULL,
    booking_info      JSONB             NOT NULL,
    payment_status    payment_status    NOT NULL DEFAULT 'pending',
    other_status      JSONB             NULL,
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
    );


CREATE INDEX booking_status_index ON bookings(payment_status);
CREATE INDEX order_status_index ON orders(payment_status);

CREATE TABLE events (
    id         SERIAL       PRIMARY KEY,
    timestamp  TIMESTAMPTZ  DEFAULT now(),
    label      VARCHAR(128),
    event_data JSONB
);

-- +migrate Down

SET ROLE vb_purchases;

DROP TABLE purchases CASCADE;
DROP TABLE orders;
DROP TABLE bookings;
DROP TABLE events;

DROP TYPE payment_status;
DROP TYPE purchase_status;

