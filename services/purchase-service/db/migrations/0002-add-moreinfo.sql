-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

ALTER TABLE orders
    ADD COLUMN order_info JSONB NOT NULL DEFAULT '{}',
    ALTER COLUMN other_status SET DEFAULT '{}',
    ALTER COLUMN other_status SET NOT NULL;

ALTER TABLE bookings
    ALTER COLUMN other_status SET DEFAULT '{}',
    ALTER COLUMN other_status SET NOT NULL;


-- +migrate Down

SET ROLE vb_purchases;

ALTER TABLE orders DROP COLUMN order_info;


