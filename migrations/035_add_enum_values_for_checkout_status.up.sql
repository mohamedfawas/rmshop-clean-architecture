ALTER TABLE checkout_sessions DROP COLUMN IF EXISTS status;

ALTER TABLE checkout_sessions
ADD COLUMN status VARCHAR(20) CHECK (status IN (
    'pending',
    'payment_initiated',
    'payment_failed',
    'completed',
    'abandoned',
    'expired'
)) DEFAULT 'pending';

CREATE INDEX idx_checkout_sessions_status ON checkout_sessions(status);