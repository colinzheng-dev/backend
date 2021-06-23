-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_payments;

SET ROLE vb_payments;

CREATE TABLE transfers (
   transfer_id       TEXT              PRIMARY KEY,
   origin            VARCHAR(24)       NOT NULL,
   destination       TEXT              NOT NULL,
   currency          VARCHAR(3)        NOT NULL,
   amount            INTEGER           NOT NULL,
   created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);


-- +migrate Down

SET ROLE vb_payments;

DROP TABLE transfers;



