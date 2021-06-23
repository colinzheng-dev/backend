-- +migrate Up

SET ROLE vb_items;

CREATE TABLE item_colls (
  id          SERIAL     PRIMARY KEY,
  name        TEXT       UNIQUE NOT NULL,
  owner       TEXT       NOT NULL
);


CREATE TABLE item_colls_items (
  coll_id  INTEGER  NOT NULL REFERENCES item_colls(id) ON DELETE CASCADE,
  idx      INTEGER  NOT NULL,
  item_id  TEXT     NOT NULL,

  PRIMARY KEY (coll_id, idx),
  UNIQUE (coll_id, item_id)
);


-- +migrate Down

SET ROLE vb_items;

DROP TABLE item_colls_items;
DROP TABLE item_colls;
DROP TYPE coll_type;
