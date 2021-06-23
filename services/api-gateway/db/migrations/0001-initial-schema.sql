-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_gateway;
SET ROLE vb_gateway;

CREATE TABLE login_tokens (
  token      CHAR(6)      PRIMARY KEY,
  email      VARCHAR(256) UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ  NOT NULL
);

CREATE INDEX login_tokens_expired_index ON login_tokens(expires_at);

-- TODO: PROBABLY NEED SOME DEVICE INFORMATION HERE TO DISTINGUISH
-- BETWEEN MULTIPLE SESSIONS FOR A SINGLE USER.
CREATE TABLE sessions (
  token    VARCHAR(32)  PRIMARY KEY,
  user_id  VARCHAR(24)  NOT NULL,
  email    VARCHAR(256) NOT NULL,
  is_admin BOOLEAN      NOT NULL DEFAULT false
);

CREATE INDEX session_email_index ON sessions(email);


-- +migrate Down

SET ROLE vb_gateway;

DROP TABLE login_tokens;
DROP TABLE sessions;
