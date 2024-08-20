CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    token TEXT PRIMARY KEY,
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);
