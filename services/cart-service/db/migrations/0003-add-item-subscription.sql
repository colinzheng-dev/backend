-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_carts;

SET ROLE vb_carts;

ALTER TABLE cart_items
    ADD COLUMN subscribe BOOLEAN DEFAULT false,
    ADD COLUMN delivery_every INTEGER NOT NULL DEFAULT 0;

-- +migrate Down

SET ROLE vb_items;

ALTER TABLE cart_items DROP COLUMN subscribe, DROP COLUMN delivery_every;

