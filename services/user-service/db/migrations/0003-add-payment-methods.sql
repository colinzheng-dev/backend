-- +migrate Up

SET ROLE vb_users;

CREATE TABLE payment_methods (
  id            VARCHAR(24)      PRIMARY KEY,
  pm_id         TEXT             NOT NULL UNIQUE,
  user_id       VARCHAR(24)      NOT NULL REFERENCES users(id), --check if an org can purchase goods
  description   TEXT             NULL,
  is_default    BOOLEAN          DEFAULT FALSE,
  type          TEXT             NOT NULL,
  other_info    JSONB            NULL,
  created_at    TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE TABLE payout_accounts (
  id            VARCHAR(24)  PRIMARY KEY,
  account       TEXT         NOT NULL,
  owner         VARCHAR(24)  UNIQUE NOT NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX payment_methods_user_idx ON payment_methods(user_id);


-- +migrate Down

SET ROLE vb_users;

DROP TABLE payment_methods;
DROP TABLE payout_accounts;
