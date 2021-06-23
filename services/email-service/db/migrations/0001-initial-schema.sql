-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_email;
SET ROLE vb_email;

CREATE TABLE topics (
  id         SERIAL       PRIMARY KEY,
  name       VARCHAR(256) NOT NULL,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);

-- +migrate Down

SET ROLE vb_email;

DROP TABLE topics;
DROP TABLE events;
