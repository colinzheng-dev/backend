-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_categories;
SET ROLE vb_categories;

CREATE TABLE categories (
  id         TEXT         PRIMARY KEY,
  label      TEXT         NOT NULL,
  schema     JSONB        NOT NULL,
  extensible BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE category_entries (
  id         SERIAL       PRIMARY KEY,
  category   TEXT         NOT NULL REFERENCES categories(id),
  label      TEXT         NOT NULL,
  lang       TEXT         NOT NULL DEFAULT 'en',
  value      JSONB        NOT NULL,
  fixed      BOOLEAN      NOT NULL DEFAULT FALSE,
  creator    TEXT         NOT NULL DEFAULT 'admin', -- REFERENCES user-service:users(id)

  UNIQUE(category, label)
);


-- +migrate Down

SET ROLE vb_categories;

DROP TABLE category_entries;
DROP TABLE categories;
DROP TABLE events;
