-- +migrate Up

SET ROLE vb_users;

ALTER TABLE orgs ADD COLUMN sso_secret TEXT;

-- +migrate Down

SET ROLE vb_users;

ALTER TABLE orgs DROP COLUMN sso_secret;

