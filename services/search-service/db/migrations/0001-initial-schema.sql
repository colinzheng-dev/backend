-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_search;
SET ROLE vb_search;

CREATE TYPE approval_status AS ENUM ('pending', 'approved','rejected');

CREATE TABLE item_full_text (
  item_id    VARCHAR(24)      PRIMARY KEY, -- FOREIGN KEY REFERENCES items(id)
  item_type  TEXT             NOT NULL,
  approval   approval_status  NOT NULL,
  full_text  TSVECTOR         NOT NULL
);

CREATE INDEX item_full_text_ft_idx ON item_full_text USING GIN(full_text);
CREATE INDEX item_full_text_item_type_index ON item_full_text(item_type);
CREATE INDEX item_full_text_approval_index ON item_full_text(approval);


CREATE TABLE item_locations (
  item_id    VARCHAR(24)       PRIMARY KEY, -- FOREIGN KEY REFERENCES items(id)
  item_type  TEXT              NOT NULL,
  approval   approval_status   NOT NULL,
  location   GEOGRAPHY(POINT)  NOT NULL
);

CREATE INDEX item_locations_spatial_idx ON item_locations USING gist(location);
CREATE INDEX item_locations_item_type_index ON item_locations(item_type);
CREATE INDEX item_locations_approval_index ON item_locations(approval);


CREATE TABLE regions (
  id        SERIAL              PRIMARY KEY,
  name      TSVECTOR            NOT NULL,
  boundary  GEOGRAPHY(POLYGON)  NOT NULL
);

CREATE INDEX regions_name_idx ON regions USING GIN(name);


CREATE TABLE events (
  id         SERIAL        PRIMARY KEY,
  timestamp  TIMESTAMPTZ   DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);


-- +migrate Down

SET ROLE vb_search;

DROP TABLE item_full_text;
DROP TABLE item_locations;
DROP TABLE regions;
DROP TABLE events;
DROP TYPE approval_status;
