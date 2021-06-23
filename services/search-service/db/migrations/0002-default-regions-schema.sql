-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_search;
SET ROLE vb_search;

-- There is a lot more columns than needed, but because we have an almost fixed and
-- small set of countries, states and provinces (a little more than 4700 summing both),
-- there is no point in cleaning the data.
CREATE TABLE default_regions (
      id serial PRIMARY KEY NOT NULL,
      region_type varchar(10) NOT NULL,
      iso_3166_2 varchar(8) NULL,
      iso_a2 varchar(3) NULL,
      postal varchar(3) NULL,
      "name" varchar(44) NULL,
      name_alt varchar(129) NULL,
      "type" varchar(38) NULL,
      type_en varchar(27) NULL,
      region varchar(43) NULL,
      latitude float8 NULL,
      longitude float8 NULL,
      adm0_a3 varchar(3) NULL,
      "admin" varchar(36) NULL,
      geounit varchar(40) NULL,
      gu_a3 varchar(3) NULL,
      name_en varchar(47) NULL,
      name_es varchar(52) NULL,
      name_de varchar(51) NULL,
      name_fr varchar(52) NULL,
      geom geometry(MULTIPOLYGON) NULL
);

CREATE INDEX default_regions_geom_idx ON default_regions USING gist (geom);

-- +migrate Down

SET ROLE vb_search;

DROP TABLE default_regions;

