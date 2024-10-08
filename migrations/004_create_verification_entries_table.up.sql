CREATE TABLE IF NOT EXISTS verification_entries (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,
    user_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    password_hash TEXT NOT NULL
);

CREATE INDEX idx_verification_entries_email ON verification_entries(email);
CREATE INDEX idx_verification_entries_expires_at ON verification_entries(expires_at);
