ALTER TABLE payments DROP COLUMN IF EXISTS payment_status;

ALTER TABLE payments
ADD COLUMN payment_status VARCHAR(20) CHECK (payment_status IN (
    'pending',
    'processing',
    'successful',
    'failed',
    'refunded',
    'partially_refunded',
    'cancelled',
    'awaiting_payment',
    'paid'
)) DEFAULT 'pending';

CREATE INDEX idx_payments_payment_status ON payments(payment_status);