-- Create OTP table
CREATE TABLE IF NOT EXISTS user_otps (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email VARCHAR(255) NOT NULL,
    otp VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_otps_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on email for faster lookups
CREATE INDEX idx_user_otps_email ON user_otps(email);

-- Create function to reset user_otps sequence
CREATE OR REPLACE FUNCTION reset_user_otps_id_seq()
RETURNS void AS $$
DECLARE
    max_id INT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM user_otps;
    EXECUTE 'ALTER SEQUENCE user_otps_id_seq RESTART WITH ' || (max_id + 1)::TEXT;
END;
$$ LANGUAGE plpgsql;