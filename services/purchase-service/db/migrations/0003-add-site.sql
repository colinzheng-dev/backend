-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

ALTER TABLE purchases
    ADD COLUMN site TEXT;


-- +migrate Down

SET ROLE vb_purchases;

ALTER TABLE purchases DROP COLUMN site;


