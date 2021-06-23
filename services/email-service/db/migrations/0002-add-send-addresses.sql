-- +migrate Up

SET ROLE vb_email;
ALTER TABLE topics ADD COLUMN send_address TEXT;


-- +migrate Down

SET ROLE vb_email;
ALTER TABLE topics DROP COLUMN send_address;
