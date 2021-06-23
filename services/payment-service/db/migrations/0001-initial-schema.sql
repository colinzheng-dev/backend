-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_payments;

SET ROLE vb_payments;

CREATE TABLE payment_intents (
   intent_id         TEXT              PRIMARY KEY,
   origin            VARCHAR(24)       NOT NULL,
   status            TEXT              NOT NULL,
   currency          VARCHAR(3)        NOT NULL,
   origin_amount     INTEGER           NOT NULL,
   created_at        TIMESTAMPTZ       NOT NULL DEFAULT now(),
   last_update       TIMESTAMPTZ       NULL
);

CREATE TABLE events (
    id         SERIAL       PRIMARY KEY,
    timestamp  TIMESTAMPTZ  DEFAULT now(),
    label      VARCHAR(128),
    event_data JSONB
);

-- +migrate Down

SET ROLE vb_payments;

DROP TABLE payment_intents;

DROP TABLE events;


