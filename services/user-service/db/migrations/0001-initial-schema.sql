-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_users;
SET ROLE vb_users;

CREATE TABLE users (
  id            VARCHAR(24)  PRIMARY KEY,
  email         VARCHAR(256) UNIQUE NOT NULL,
  name          VARCHAR(256),
  display_name  VARCHAR(128),
  avatar        VARCHAR(256),
  country       CHAR(2),
  is_admin      BOOLEAN      NOT NULL DEFAULT false,
  last_login    TIMESTAMPTZ  NOT NULL,
  api_key       VARCHAR(64)
);

CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);

-- +migrate Down

SET ROLE vb_users;

DROP TABLE users;
DROP TABLE events;
