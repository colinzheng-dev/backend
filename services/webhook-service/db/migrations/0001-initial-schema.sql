-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_webhooks;
SET ROLE vb_webhooks;

CREATE TABLE webhooks
(
    id         VARCHAR(20) PRIMARY KEY,
    owner      VARCHAR(20) NOT NULL,
    url        TEXT        NOT NULL,
    enabled    BOOLEAN     NOT NULL DEFAULT FALSE,
    livemode   BOOLEAN     NOT NULL DEFAULT FALSE,
    events     TEXT[]      NOT NULL,
    secret     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX owner_webhook_idx ON webhooks (owner);

CREATE TABLE event_types
(
    name        TEXT PRIMARY KEY,
    category    TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE TABLE events
(
    event_id      TEXT PRIMARY KEY,
    destination   TEXT        NOT NULL,
    type          TEXT        NOT NULL,
    livemode      BOOLEAN     NOT NULL DEFAULT FALSE,
    sent          BOOLEAN     NOT NULL DEFAULT FALSE,
    sent_at       TIMESTAMPTZ,
    attempts      INTEGER     NOT NULL DEFAULT 0,
    retry         BOOLEAN     NOT NULL DEFAULT TRUE,
    last_retry    TIMESTAMPTZ,
    backoff_until TIMESTAMPTZ,
    payload       JSON        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX owner_events_idx ON events (destination);

-- +migrate Down

SET ROLE vb_webhooks;

DROP TABLE webhooks;
DROP TABLE event_types;
DROP TABLE events;


