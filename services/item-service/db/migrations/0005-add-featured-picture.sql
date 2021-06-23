-- +migrate Up

SET ROLE vb_items;

ALTER TABLE items ADD COLUMN featured_picture TEXT;
UPDATE items SET featured_picture = pictures[1];
ALTER TABLE items ALTER COLUMN featured_picture SET NOT NULL;


-- +migrate Down

SET ROLE vb_items;

ALTER TABLE items DROP COLUMN featured_picture;
