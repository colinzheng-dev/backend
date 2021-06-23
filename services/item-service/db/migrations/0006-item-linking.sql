-- +migrate Up

SET ROLE vb_items;

CREATE TYPE link_ownership_class AS ENUM (
  'owner-to-owner', 'owner-to-any', 'any-to-owner'
);

CREATE TABLE item_link_types (
  name           TEXT                  PRIMARY KEY,
  origin_type    TEXT[],
  target_type    TEXT[],
  origin_unique  BOOLEAN               NOT NULL DEFAULT TRUE,
  ownership      link_ownership_class  NOT NULL DEFAULT 'owner-to-owner',
  is_inverse     BOOLEAN               NOT NULL DEFAULT FALSE,
  inverse        TEXT
);


INSERT INTO item_link_types
 (name,
  origin_type, target_type, origin_unique, ownership,
  is_inverse, inverse)
VALUES
 ('is-room-in-hotel',
  '{"room"}', '{"hotel"}', TRUE, 'owner-to-owner',
  FALSE, 'hotel-has-rooms'),
 ('hotel-has-rooms',
  '{"hotel"}', '{"room"}', FALSE, 'owner-to-owner',
  TRUE, 'is-room-in-hotel'),

 ('is-offered-by',
  '{"offer"}', '{"hotel","restaurant","cafe","shop","room"}', TRUE, 'owner-to-owner',
  FALSE, 'has-offers'),
 ('has-offers',
  '{"hotel","restaurant","cafe","shop","room"}', '{"offer"}', FALSE, 'owner-to-owner',
  TRUE, 'is-offered-by'),

 ('is-media-for',
  '{"article","recipe","jobad","post"}', '{}', FALSE, 'any-to-owner',
  FALSE, 'has-linked-media'),
 ('has-linked-media',
  '{}', '{"article","recipe","jobad","post"}', FALSE, 'owner-to-any',
  TRUE, 'is-media-for'),

 ('is-offering-for',
  '{"prodffering"}', '{"packagedfood","freshfood","fashion","cosmetics","homeware"}', TRUE, 'owner-to-any',
  FALSE, 'product-has-offerings'),
 ('product-has-offerings',
  '{"packagedfood","freshfood","fashion","cosmetics","homeware"}', '{"prodffering"}', FALSE, 'any-to-owner',
  TRUE, 'is-offering-for'),

 ('offering-is-sold-in',
  '{"prodoffering"}', '{"shop"}', FALSE, 'owner-to-owner',
  FALSE, 'shop-sells'),
 ('shop-sells',
  '{"shop"}', '{"prodoffering"}', FALSE, 'owner-to-owner',
  TRUE, 'offering-is-sold-in');


CREATE TABLE item_links (
  id          VARCHAR(24)  PRIMARY KEY, -- lnk_<random>
  inverse_id  VARCHAR(24)  REFERENCES item_links(id) ON DELETE CASCADE,
  origin      VARCHAR(24)  NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  target      VARCHAR(24)  NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  link_type   TEXT         NOT NULL REFERENCES item_link_types(name) ON DELETE CASCADE,
  owner       TEXT         NOT NULL, -- REFERENCES users(id)
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);


-- +migrate Down

SET ROLE vb_items;

DROP TABLE item_links;
DROP TABLE item_link_types;
DROP TYPE link_ownership_class;
