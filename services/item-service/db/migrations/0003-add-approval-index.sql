-- +migrate Up

SET ROLE vb_items;
CREATE INDEX item_approval_index ON items(approval);


-- +migrate Down

SET ROLE vb_items;
DROP INDEX item_approval_index;
