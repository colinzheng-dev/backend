-- +migrate Up

SET ROLE vb_users;


CREATE TABLE delivery_fees
(
    id                  VARCHAR(24) PRIMARY KEY,
    owner               VARCHAR(24) UNIQUE NOT NULL,
    free_delivery_above INTEGER            NOT NULL,
    normal_order_price  INTEGER            NOT NULL,
    chilled_order_price INTEGER            NOT NULL,
    currency            TEXT               NOT NULL,
    created_at          TIMESTAMPTZ        NOT NULL DEFAULT now()
);

CREATE INDEX del_fee_user_idx ON delivery_fees (owner);

-- +migrate Down

SET ROLE vb_users;

DROP TABLE delivery_fees;

