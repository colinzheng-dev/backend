-- +migrate Up

SET ROLE vb_sites;
ALTER TABLE sites ADD COLUMN fee NUMERIC(5, 4) NOT NULL DEFAULT 0.15;


-- +migrate Down

SET ROLE vb_sites;
ALTER TABLE sites DROP COLUMN fee;
