-- +migrate Up

SET ROLE vb_gateway;

ALTER TABLE login_tokens
  ADD COLUMN site       VARCHAR(24),
  ADD COLUMN language   VARCHAR(2);

CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);


-- +migrate Down

SET ROLE vb_gateway;

ALTER TABLE login_tokens
  DROP COLUMN site,
  DROP COLUMN language;

DROP TABLE events;
