INSERT INTO login_tokens (token, email, site, language, expires_at)
  VALUES ('123456', 'user@test.com', 'veganbase', 'en', now() + INTERVAL '1 day');
INSERT INTO login_tokens (token, email, site, language, expires_at)
  VALUES ('654321', 'user2@test.com', 'veganbase', 'en', now() + INTERVAL '1 day');

INSERT INTO sessions VALUES ('SESSION-1', 'usr_TESTUSER1', 'test1@example.com', false);
INSERT INTO sessions VALUES ('SESSION-2A', 'usr_TESTUSER2', 'test2@example.com', false);
INSERT INTO sessions VALUES ('SESSION-2B', 'usr_TESTUSER2', 'test2@example.com', false);
INSERT INTO sessions VALUES ('SESSION-3', 'usr_TESTUSER3', 'test3@example.com', true);
