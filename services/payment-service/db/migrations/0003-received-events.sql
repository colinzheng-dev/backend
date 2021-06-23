-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_payments;

SET ROLE vb_payments;

CREATE TABLE received_events
(
    event_id        TEXT PRIMARY KEY,
    idempotency_key TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    is_handled      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE error_logs
(
    event_id   TEXT PRIMARY KEY,
    error      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transfer_remainders
(
    transfer_id           TEXT PRIMARY KEY,
    destination           TEXT          NOT NULL,
    currency              VARCHAR(3)    NOT NULL,
    total_value        INTEGER       NOT NULL,
    fee_value             INTEGER       NOT NULL,
    transferred_value     INTEGER       NOT NULL,
    fee_remainder         NUMERIC(5, 4) NOT NULL, --FRACTION OF CENTS
    transferred_remainder NUMERIC(5, 4) NOT NULL, --FRACTION OF CENTS
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT now()
);
-- +migrate Down

SET ROLE vb_payments;

DROP TABLE received_events;
DROP TABLE transfer_remainders;
DROP TABLE error_logs;



