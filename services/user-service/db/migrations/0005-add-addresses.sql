-- +migrate Up

SET ROLE vb_users;

CREATE TABLE addresses
(
    id             VARCHAR(24) PRIMARY KEY,
    owner          VARCHAR(24) NOT NULL REFERENCES users (id),
    description    TEXT        NOT NULL,
    street_address TEXT        NOT NULL,
    city           TEXT        NOT NULL,
    postcode       TEXT        NOT NULL,
    country        TEXT        NOT NULL,
    house_number   TEXT        NOT NULL,
    is_default     BOOLEAN     NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- +migrate Down

SET ROLE vb_users;

DROP TABLE addresses;

