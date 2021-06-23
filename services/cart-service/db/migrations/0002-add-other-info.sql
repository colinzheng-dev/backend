-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_carts;

SET ROLE vb_carts;

ALTER TABLE cart_items
    ADD COLUMN item_type TEXT,
    ADD COLUMN other_info JSONB NOT NULL DEFAULT '{}';

ALTER TABLE cart_items DROP CONSTRAINT cart_items_cart_id_item_id_key;

-- +migrate Down

SET ROLE vb_items;

ALTER TABLE cart_items DROP COLUMN item_type, DROP COLUMN other_info;

