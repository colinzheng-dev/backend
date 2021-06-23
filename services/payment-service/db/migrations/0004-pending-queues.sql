-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_payments;

SET ROLE vb_payments;

ALTER TABLE transfers ADD COLUMN destination_account TEXT NOT NULL DEFAULT '';
ALTER TABLE transfer_remainders ADD COLUMN destination_account TEXT NOT NULL default '';

CREATE TABLE pending_events
(
    event_id    TEXT PRIMARY KEY,
    intent_id   TEXT        NOT NULL,
    reason      TEXT        NOT NULL,
    attempts    INTEGER     NOT NULL DEFAULT 0,
    last_update TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pending_transfers
(
    id                    BIGSERIAL PRIMARY KEY,
    origin                VARCHAR(24)   NOT NULL,
    destination           TEXT          NOT NULL,
    currency              VARCHAR(3)    NOT NULL,
    source_transaction    TEXT          NOT NULL,
    total_value           INTEGER       NOT NULL,
    fee_value             INTEGER       NOT NULL,
    transferred_value     INTEGER       NOT NULL,
    fee_remainder         NUMERIC(5, 4) NOT NULL, --FRACTION OF CENTS
    transferred_remainder NUMERIC(5, 4) NOT NULL, --FRACTION OF CENTS
    reason                TEXT          NOT NULL,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- +migrate Down

SET ROLE vb_payments;

DROP TABLE pending_events;
DROP TABLE pending_transfers;

ALTER TABLE transfers DROP COLUMN destination_account;
ALTER TABLE transfer_remainders DROP COLUMN destination_account;
