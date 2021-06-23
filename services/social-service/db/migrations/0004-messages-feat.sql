-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

SET ROLE vb_social;

CREATE TYPE thread_status AS ENUM ('open', 'closed', 'archived', 'deleted');

CREATE TABLE threads
(
    id           VARCHAR(24) PRIMARY KEY,
    subject      TEXT        NOT NULL,
    author       VARCHAR(24) NOT NULL,
    content      TEXT        NOT NULL,
    attachments  JSONB,
    lock_reply   BOOLEAN     NOT NULL DEFAULT FALSE,
    participants TEXT [],
    status       TEXT        NOT NULL DEFAULT 'open',
    is_edited    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX threads_author_index ON threads (author);

CREATE TABLE messages
(
    id          VARCHAR(24) PRIMARY KEY,
    parent_id   VARCHAR(24) NOT NULL REFERENCES threads (id) ON DELETE CASCADE,
    author      VARCHAR(24) NOT NULL,
    content     TEXT        NOT NULL,
    attachments JSONB,
    is_edited   BOOLEAN     NOT NULL DEFAULT FALSE,
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX messages_parentid_index ON messages (parent_id);


-- +migrate Down

SET ROLE vb_social;

DROP TABLE threads;
DROP TABLE messages;

