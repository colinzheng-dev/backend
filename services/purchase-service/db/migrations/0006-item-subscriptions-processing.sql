-- +migrate Up

ALTER TABLE gorp_migrations OWNER TO vb_purchases;

SET ROLE vb_purchases;

CREATE
TYPE processing_status AS ENUM ('pending', 'processing', 'completed', 'error');

CREATE TABLE subscription_purchases_processing
(
    id                SERIAL PRIMARY KEY,
    is_processing_day BOOL              NOT NULL DEFAULT FALSE,
    reference         DATE              NOT NULL,
    status            processing_status NOT NULL DEFAULT 'pending',
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT now(),
    started_at        TIMESTAMPTZ       NULL,
    ended_at          TIMESTAMPTZ       NULL
);

--creates a processing entry for each day on the next 10 years.
INSERT INTO subscription_purchases_processing (reference)
SELECT ser.day::DATE as reference
from generate_series(now(),
                     to_date(concat((EXTRACT(YEAR FROM NOW()) + 10):: text, '1231'),
                             'yyyyMMdd'), '1 day') AS ser(day);

CREATE TABLE subscription_purchases
(
    id           SERIAL PRIMARY KEY,
    reference    DATE            NOT NULL,
    buyer_id     VARCHAR(24)     NOT NULL, -- FOREIGN KEY REFERENCES vb_users/users(id)
    address_id   VARCHAR(24)     NOT NULL,
    status       purchase_status NOT NULL DEFAULT 'pending',
    purchase_id  VARCHAR(24)     NULL,
    errors       TEXT            NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ     NULL
);



CREATE INDEX subscription_purchases_processing_date_index ON subscription_purchases_processing (reference);
CREATE INDEX subscription_purchases_date_index ON subscription_purchases (reference);

-- +migrate Down

SET ROLE vb_purchases;

DROP TABLE subscription_purchases_processing;
DROP TABLE subscription_purchases;
DROP
TYPE processing_status;



