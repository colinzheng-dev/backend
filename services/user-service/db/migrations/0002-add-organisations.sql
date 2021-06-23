-- +migrate Up

SET ROLE vb_users;

CREATE TABLE orgs (
  id            VARCHAR(24)      PRIMARY KEY,
  name          TEXT             NOT NULL,
  slug          TEXT             UNIQUE NOT NULL,
  logo          TEXT,
  description   TEXT,
  address       JSONB,
  phone         TEXT,
  email         TEXT,
  urls          JSONB,
  industry      TEXT[],
  year_founded  INTEGER,
  employees     INTEGER,
  created_at    TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE TABLE org_users (
  id            SERIAL       PRIMARY KEY,
  org_id        VARCHAR(24)  NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  user_id       VARCHAR(24)  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_org_admin  BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),

  UNIQUE(org_id, user_id)
);

CREATE INDEX org_user_user_idx ON org_users(user_id);


-- +migrate Down

SET ROLE vb_users;

DROP TABLE org_user_invitations;
DROP TABLE org_users;
DROP TABLE orgs;
