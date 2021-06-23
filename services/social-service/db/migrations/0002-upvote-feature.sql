-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

SET ROLE vb_social;

CREATE TABLE upvotes (
  upvote_id         VARCHAR(24)  PRIMARY KEY,
  user_id           VARCHAR(24)  NOT NULL,
  item_id           VARCHAR(24)  NOT NULL,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, item_id)
);

CREATE TABLE events (
                        id         SERIAL       PRIMARY KEY,
                        timestamp  TIMESTAMPTZ  DEFAULT now(),
                        label      VARCHAR(128),
                        event_data JSONB
);
-- +migrate Down

SET ROLE vb_social;

DROP TABLE upvotes;
DROP TABLE events;
