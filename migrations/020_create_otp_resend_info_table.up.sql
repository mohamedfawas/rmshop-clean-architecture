CREATE TABLE otp_resend_info (
    email VARCHAR(255) PRIMARY KEY,
    resend_count INT NOT NULL DEFAULT 0,
    last_resend_time TIMESTAMP WITH TIME ZONE NOT NULL
);