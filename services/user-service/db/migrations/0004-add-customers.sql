-- +migrate Up

SET ROLE vb_users;

CREATE TABLE customers (
  user_id       VARCHAR(24)      UNIQUE NOT NULL REFERENCES users(id), --check if an org can purchase goods
  customer_id   TEXT             PRIMARY KEY,
  created_at    TIMESTAMPTZ      NOT NULL DEFAULT now()
);


-- +migrate Down

SET ROLE vb_users;

DROP TABLE customers;

