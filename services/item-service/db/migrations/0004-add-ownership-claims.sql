-- +migrate Up

SET ROLE vb_items;

CREATE TABLE ownership_claims (
  id          VARCHAR(24)      PRIMARY KEY,
  user_id     VARCHAR(24)      NOT NULL,
  item_id     VARCHAR(24)      REFERENCES items(id) ON DELETE CASCADE,
  status      approval_status  NOT NULL DEFAULT 'pending',
  created_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX claim_user_index ON ownership_claims(user_id);
CREATE INDEX claim_status_index ON ownership_claims(status);
CREATE INDEX claim_created_at_index ON ownership_claims(created_at);


-- +migrate Down

SET ROLE vb_items;

DROP TABLE ownership_claims;
