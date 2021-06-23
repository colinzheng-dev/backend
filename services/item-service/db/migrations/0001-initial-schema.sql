-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_items;
SET ROLE vb_items;


-- Queries we want to be able to do:
--
--  * Approved item IDs by type, ordered by creation date.
--  * Approved item IDs by tag, ordered by creation date.
--  * Filter by approval status.
--  * Item IDs by owner.
--  * Item IDs by some attributes?

-- To be implemented by separate search service:
--
--  * Full-text search on name, description, tags, some attributes.
--  * Location search as distance from point.
--  * Location search within a region.
--  * Geographical search (name of region -> search within region).

-- Views we want:
--
--  * Summary view: id, item_type, slug, lang, name, description,
--    featured_picture, urls, owner
--  * Detail view: summary view + pictures, attrs
--  * Admin view: detail view + approval, creator, ownership


CREATE TYPE approval_status AS ENUM ('pending', 'approved','rejected');
CREATE TYPE ownership_status AS ENUM ('creator', 'claimed');

CREATE TABLE items (
  id                VARCHAR(24)       PRIMARY KEY,
  item_type         TEXT              NOT NULL,
  slug              TEXT              UNIQUE NOT NULL,
  lang              TEXT              NOT NULL DEFAULT 'en',
  name              TEXT              NOT NULL,
  description       TEXT,
  featured_picture  TEXT,
  pictures          TEXT[],
  tags              TEXT[],
  urls              JSONB,
  attrs             JSONB,
  approval          approval_status   NOT NULL DEFAULT 'pending',
  creator           TEXT              NOT NULL, -- FOREIGN KEY REFERENCES vb_users/users(id)
  owner             TEXT              NOT NULL, -- FOREIGN KEY REFERENCES vb_users/users(id)
  ownership         ownership_status  NOT NULL DEFAULT 'creator',
  created_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE INDEX item_type_index ON items(item_type);
CREATE INDEX item_slug_index ON items(slug);
CREATE INDEX item_tag_index ON items USING GIN(tags);
CREATE INDEX item_attrs_index ON items USING GIN(attrs);
CREATE INDEX item_created_at_index ON items(created_at);


CREATE TABLE events (
  id         SERIAL       PRIMARY KEY,
  timestamp  TIMESTAMPTZ  DEFAULT now(),
  label      VARCHAR(128),
  event_data JSONB
);

-- +migrate Down

SET ROLE vb_items;

DROP TABLE items;
DROP TABLE events;
DROP TYPE approval_status;
DROP TYPE ownership_status;
