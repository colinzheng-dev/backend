-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_sites;
SET ROLE vb_sites;

CREATE TABLE sites (
  id         VARCHAR(24)  PRIMARY KEY,
  name       VARCHAR(256) NOT NULL,
  url        VARCHAR(256) NOT NULL,
  signature  VARCHAR(256) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);

-- +migrate Down

SET ROLE vb_sites;

DROP TABLE sites;
DROP TABLE events;
