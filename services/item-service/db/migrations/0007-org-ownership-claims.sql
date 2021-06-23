-- +migrate Up

SET ROLE vb_items;

ALTER TABLE ownership_claims RENAME COLUMN user_id TO owner_id;


-- +migrate Down

SET ROLE vb_items;

ALTER TABLE ownership_claims RENAME COLUMN owner_id TO user_id;
