-- +migrate Up

SET ROLE vb_users;

ALTER TABLE users ADD COLUMN secret_key TEXT;
ALTER TABLE users ALTER COLUMN api_key TYPE TEXT;
ALTER TABLE users ADD CONSTRAINT unique_api_keys UNIQUE (api_key);

-- +migrate Down
SET ROLE vb_users;

ALTER TABLE users DROP COLUMN secret_key;
ALTER TABLE users ALTER COLUMN api_key TYPE varchar(64);
ALTER TABLE users DROP CONSTRAINT unique_api_keys;


