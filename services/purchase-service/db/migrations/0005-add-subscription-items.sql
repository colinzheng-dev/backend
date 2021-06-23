-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

CREATE
TYPE subscription_status AS ENUM ('active', 'paused', 'deleted');

CREATE TABLE subscription_items
(
    id             VARCHAR(24) PRIMARY KEY,
    owner          VARCHAR(24)         NOT NULL,
    item_id        VARCHAR(24)         NOT NULL,
    item_type      VARCHAR(24)         NOT NULL,
    address_id     VARCHAR(24)         NULL,
    origin         VARCHAR(24)         NOT NULL REFERENCES PURCHASES (id) ON DELETE CASCADE,
    quantity       INTEGER             NOT NULL,
    other_info     JSONB               NOT NULL DEFAULT '{}',
    delivery_every INTEGER             NOT NULL DEFAULT 0,
    status         subscription_status NOT NULL DEFAULT 'active',
    next_delivery  INTEGER             NOT NULL DEFAULT 0,
    last_delivery  TIMESTAMPTZ         NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ         NOT NULL DEFAULT now(),
    active_since   TIMESTAMPTZ         NULL,
    paused_since   TIMESTAMPTZ         NULL,
    deleted_at     TIMESTAMPTZ         NULL
);

CREATE INDEX subscription_items_owner_index ON subscription_items (owner);

-- +migrate Down

SET ROLE vb_purchases;

DROP TABLE subscription_items;
DROP TYPE subscription_status;



