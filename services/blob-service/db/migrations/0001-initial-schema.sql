-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_blobs;
SET ROLE vb_blobs;

/* Utility to allow aggregation of JSONB objects. */
CREATE AGGREGATE jsonb_object_agg(jsonb) (
   SFUNC = 'jsonb_concat',
   STYPE = jsonb,
   INITCOND = '{}'
);

/*
   The owner column here can be NULL is a user uploads a blob, assigns
   it to an item, then deletes the blob from their image gallery. The
   blob itself cannot be deleted because it is still in use by items.
*/

CREATE TABLE blobs (
  id         VARCHAR(24)  PRIMARY KEY,
  uri        VARCHAR(256) NOT NULL,
  format     VARCHAR(64)  NOT NULL,     /* MIME type. */
  size       INTEGER      NOT NULL,
  owner      VARCHAR(256),              /* User ID. */
  tags       JSONB,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

/* We need to be able to retrieve a user's blobs. */
CREATE INDEX blob_owner_index ON blobs(owner, created_at);

/*
   Association between blobs and items. Blobs are not deleted until
   they are not associated with any items *and* the user has deleted
   the blob from their image gallery.
*/

CREATE TABLE blob_items (
  id      SERIAL       PRIMARY KEY,
  blob_id VARCHAR(256) NOT NULL REFERENCES blobs(id) ON DELETE CASCADE,
  item_id VARCHAR(256) NOT NULL,

  UNIQUE(blob_id, item_id)
);

/* We need to be able to retrieve blobs from items and items from
   blobs. */
CREATE INDEX blob_item_blob_id_index ON blob_items(blob_id);
CREATE INDEX blob_item_item_id_index ON blob_items(item_id);


CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);

-- +migrate Down

SET ROLE vb_blobs;

DROP TABLE blob_items;
DROP TABLE blob_tags;
DROP TABLE blobs;
DROP TABLE events;
