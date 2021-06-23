-- +migrate Up

SET ROLE vb_items;

CREATE TABLE item_statistics (
  item_id VARCHAR(24) PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
  upvotes INTEGER NOT NULL,
  rank FLOAT NOT NULL
);


-- +migrate Down

SET ROLE vb_items;

DROP TABLE item_statistics;

