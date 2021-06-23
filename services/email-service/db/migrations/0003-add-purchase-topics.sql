-- +migrate Up

SET ROLE vb_email;

INSERT INTO topics (name, send_address, created_at) VALUES
    ('purchase-created-topic', 'hello', now()),
    ('order-created-topic', 'hello', now()),
    ('booking-created-topic', 'hello', now()),
    ('payment-received-topic', 'hello', now());


-- +migrate Down

SET ROLE vb_email;
DELETE FROM topics WHERE name IN
    ('purchase-created-topic', 'order-created-topic', 'booking-created-topic', 'payment-received-topic');
