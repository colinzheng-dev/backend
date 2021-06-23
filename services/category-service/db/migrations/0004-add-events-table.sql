-- +migrate Up

SET ROLE vb_categories;

CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);


-- +migrate Down

SET ROLE vb_categories;

DROP TABLE events;
