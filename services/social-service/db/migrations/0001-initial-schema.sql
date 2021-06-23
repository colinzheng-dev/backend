-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_social;
SET ROLE vb_social;

CREATE TABLE subscriptions (
  user_id           VARCHAR(24)  NOT NULL,
  subscription_id   VARCHAR(256) NOT NULL,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, subscription_id)
);

-- +migrate Down

SET ROLE vb_social;

DROP TABLE subscriptions;
