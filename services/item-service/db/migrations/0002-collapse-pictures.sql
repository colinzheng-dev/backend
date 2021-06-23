-- +migrate Up

SET ROLE vb_items;

UPDATE items SET pictures = featured_picture || pictures;
ALTER TABLE items DROP COLUMN featured_picture;


-- +migrate Down

SET ROLE vb_items;

ALTER TABLE items
  ADD COLUMN featured_picture TEXT;
UPDATE items
   SET featured_picture = pictures[1],
       pictures = pictures[2:];
