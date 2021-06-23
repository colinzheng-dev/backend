-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

ALTER TABLE purchases
    ADD COLUMN delivery_fees JSONB DEFAULT '[]';

ALTER TABLE orders
    ADD COLUMN delivery_fee JSONB DEFAULT '{}';

-- +migrate Down

SET ROLE vb_purchases;

ALTER TABLE purchases DROP COLUMN delivery_fees;
ALTER TABLE orders DROP COLUMN delivery_fee;


