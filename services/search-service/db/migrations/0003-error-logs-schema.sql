-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_search;
SET ROLE vb_search;

CREATE TABLE error_logs
(
    ID         SERIAL PRIMARY KEY,
    action     TEXT        NOT NULL DEFAULT 'generic',
    error      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- +migrate Down

SET ROLE vb_search;

DROP TABLE error_logs;

