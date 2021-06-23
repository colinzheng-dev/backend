-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up

SET ROLE vb_social;

CREATE TABLE posts
(
    id         VARCHAR(24) PRIMARY KEY,
    post_type  VARCHAR(24) NOT NULL,
    owner      VARCHAR(24) NOT NULL,
    subject    VARCHAR(24) NOT NULL,
    is_edited  BOOLEAN     NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN     NOT NULL DEFAULT FALSE,
    pictures   TEXT [],
    attrs      JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX post_subject_index ON posts (subject);
CREATE INDEX post_type_index ON posts (post_type);

CREATE TABLE replies
(
    id         VARCHAR(24) PRIMARY KEY,
    parent_id  VARCHAR(24) NOT NULL,
    owner      VARCHAR(24) NOT NULL,
    is_edited  BOOLEAN     NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN     NOT NULL DEFAULT FALSE,
    pictures   TEXT [],
    attrs      JSONB       NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX reply_parentid_index ON replies (parent_id);


-- +migrate Down

SET ROLE vb_social;

DROP TABLE posts;
DROP TABLE replies;

