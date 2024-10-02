ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_order_status_check;

ALTER TABLE orders
    ADD CONSTRAINT orders_order_status_check CHECK (order_status IN (
        'pending_payment',
        'processing',
        'shipped',
        'completed',
        'cancelled',
        'refunded',
        'confirmed',
        'pending_cancellation',
        'return_approved'
    ));
