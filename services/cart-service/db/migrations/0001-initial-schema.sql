-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_carts;

SET ROLE vb_carts;

CREATE TYPE cart_status AS ENUM ('active', 'complete', 'abandoned', 'not logged in');

CREATE TABLE carts (
  id                VARCHAR(24)      PRIMARY KEY,
  cart_status       TEXT              NOT NULL,
  owner             TEXT              NULL, -- FOREIGN KEY REFERENCES vb_users/users(id)
  created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE TABLE cart_items (
  id      SERIAL  PRIMARY KEY,
  cart_id  VARCHAR(24)  NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
  item_id  TEXT     NOT NULL,
  quantity  INTEGER  NOT NULL,
  UNIQUE (cart_id, item_id)
);

CREATE INDEX cart_status_index ON carts(cart_status);
CREATE INDEX cart_owner_index ON carts(owner);


CREATE TABLE events (
                        id         SERIAL       PRIMARY KEY,
                        timestamp  TIMESTAMPTZ  DEFAULT now(),
                        label      VARCHAR(128),
                        event_data JSONB
);

-- +migrate Down

SET ROLE vb_items;

DROP TABLE events;
DROP TABLE carts;
DROP TABLE cart_items;
DROP TYPE cart_status;

