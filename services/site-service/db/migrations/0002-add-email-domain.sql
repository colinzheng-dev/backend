-- +migrate Up

SET ROLE vb_sites;
ALTER TABLE sites ADD COLUMN email_domain TEXT;
UPDATE sites SET email_domain = replace(replace(url, 'https://', ''), 'http://', '');
ALTER TABLE sites ALTER COLUMN email_domain SET NOT NULL;


-- +migrate Down

SET ROLE vb_sites;
ALTER TABLE sites DROP COLUMN email_domain;
