-- +migrate Up

SET ROLE vb_users;

ALTER TABLE addresses ADD COLUMN region_postal TEXT DEFAULT '';
ALTER TABLE addresses ADD COLUMN recipient JSON DEFAULT '{}';
ALTER TABLE addresses ADD COLUMN coordinates JSON DEFAULT '{}';

-- +migrate Down
SET ROLE vb_users;

ALTER TABLE addresses DROP COLUMN region_postal;
ALTER TABLE addresses DROP COLUMN recipient;
ALTER TABLE addresses DROP COLUMN coordinates;

